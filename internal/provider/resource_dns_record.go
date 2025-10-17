package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure resource implementation
var _ resource.Resource = &dnsRecordResource{}
var _ resource.ResourceWithImportState = &dnsRecordResource{}

func NewDNSRecordResource() resource.Resource {
	return &dnsRecordResource{}
}

type dnsRecordResource struct {
	client *Client
}

type dnsRecordModel struct {
	ID       types.String `tfsdk:"id"`
	Service  types.String `tfsdk:"service"`
	Domain   types.String `tfsdk:"domain"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Content  types.String `tfsdk:"content"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Priority types.Int64  `tfsdk:"priority"`
}

func (r *dnsRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *dnsRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "Domain name owning the record (zone)",
			},
			"service": schema.StringAttribute{
				Optional:    true,
				Description: "Active24 v2 service key (if different from domain). If set, this overrides domain in API path /v2/service/{service}",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Record name (relative or FQDN depending on API)",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "DNS record type (A, AAAA, CNAME, TXT, MX, etc.)",
			},
			"content": schema.StringAttribute{
				Required: true,
			},
			"ttl": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(3600),
			},
			"priority": schema.Int64Attribute{
				Optional:    true,
				Description: "Priority for MX/SRV where applicable",
			},
		},
	}
}

func (r *dnsRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsRecordModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := createRecordRequest{
		Name:    normalizeNameForAPI(plan.Name.ValueString()),
		Type:    plan.Type.ValueString(),
		Content: plan.Content.ValueString(),
		TTL:     plan.TTL.ValueInt64(),
	}
	if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
		p := plan.Priority.ValueInt64()
		createReq.Priority = &p
	}

	targetService := plan.Domain.ValueString()
	if !plan.Service.IsNull() && !plan.Service.IsUnknown() && plan.Service.ValueString() != "" {
		targetService = plan.Service.ValueString()
	}

	_, err := r.client.CreateRecord(ctx, targetService, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating record", err.Error())
		return
	}

	// After create, API may return 204 with no body. Fetch the record to obtain ID/TTL.
	records, err := r.client.ListRecords(ctx, targetService, normalizeNameForAPI(plan.Name.ValueString()), plan.Type.ValueString(), plan.Content.ValueString(), ptrI(plan.TTL.ValueInt64()))
	if err != nil || len(records) == 0 {
		resp.Diagnostics.AddError("Error reading created record", fmt.Sprintf("lookup failed: %v", err))
		return
	}
	rec := records[0]

	plan.ID = types.StringValue(fmt.Sprintf("%d", rec.ID))
	plan.TTL = types.Int64Value(rec.TTL)
	if rec.Priority != nil {
		plan.Priority = types.Int64Value(*rec.Priority)
	} else {
		plan.Priority = types.Int64Null()
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsRecordModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	targetService := state.Domain.ValueString()
	if !state.Service.IsNull() && !state.Service.IsUnknown() && state.Service.ValueString() != "" {
		targetService = state.Service.ValueString()
	}
	// v2: list and match by ID
	records, err := r.client.ListRecords(ctx, targetService, normalizeNameForAPI(state.Name.ValueString()), state.Type.ValueString(), "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading record", err.Error())
		return
	}
	var rec *DNSRecord
	for i := range records {
		if records[i].ID == id {
			rec = &records[i]
			break
		}
	}
	if rec == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(denormalizeNameFromAPI(rec.Name))
	state.Type = types.StringValue(rec.Type)
	state.Content = types.StringValue(rec.Content)
	state.TTL = types.Int64Value(rec.TTL)
	if rec.Priority != nil {
		state.Priority = types.Int64Value(*rec.Priority)
	} else {
		state.Priority = types.Int64Null()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *dnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsRecordModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	updateReq := createRecordRequest{
		Name:    normalizeNameForAPI(plan.Name.ValueString()),
		Type:    plan.Type.ValueString(),
		Content: plan.Content.ValueString(),
		TTL:     plan.TTL.ValueInt64(),
	}
	if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
		p := plan.Priority.ValueInt64()
		updateReq.Priority = &p
	}

	targetService := plan.Domain.ValueString()
	if !plan.Service.IsNull() && !plan.Service.IsUnknown() && plan.Service.ValueString() != "" {
		targetService = plan.Service.ValueString()
	}
	if _, err := r.client.UpdateRecord(ctx, targetService, id, updateReq); err != nil {
		resp.Diagnostics.AddError("Error updating record", err.Error())
		return
	}

	// Re-read to reflect server TTL normalization
	records, err := r.client.ListRecords(ctx, targetService, normalizeNameForAPI(plan.Name.ValueString()), plan.Type.ValueString(), plan.Content.ValueString(), ptrI(plan.TTL.ValueInt64()))
	if err == nil && len(records) > 0 {
		rec := records[0]
		plan.TTL = types.Int64Value(rec.TTL)
		if rec.Priority != nil {
			plan.Priority = types.Int64Value(*rec.Priority)
		} else {
			plan.Priority = types.Int64Null()
		}
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *dnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsRecordModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	targetService := state.Domain.ValueString()
	if !state.Service.IsNull() && !state.Service.IsUnknown() && state.Service.ValueString() != "" {
		targetService = state.Service.ValueString()
	}
	if err := r.client.DeleteRecord(ctx, targetService, id); err != nil {
		resp.Diagnostics.AddError("Error deleting record", err.Error())
		return
	}
}

func (r *dnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expect import ID in the form: domain:id
	// Example: example.com:12345
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import format", "Expected '<domain>:<id>'")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// normalizeNameForAPI converts Terraform-friendly apex marker "@" to the API's expected apex representation.
func normalizeNameForAPI(name string) string {
	if name == "@" {
		return ""
	}
	return name
}

// denormalizeNameFromAPI converts API apex representation back to Terraform-friendly "@".
func denormalizeNameFromAPI(name string) string {
	if name == "" {
		return "@"
	}
	return name
}

func ptrI(v int64) *int64 { return &v }
