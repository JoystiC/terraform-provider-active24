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
	CAAValue types.String `tfsdk:"caa_value"`
	CAAFlags types.Int64  `tfsdk:"caa_flags"`
	CAATag   types.String `tfsdk:"caa_tag"`
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
				Description: "DNS record type (A, AAAA, CNAME, TXT, MX, CAA, etc.)",
			},
			"content": schema.StringAttribute{
				Optional:    true,
				Description: "Record content (not used for CAA)",
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
			"caa_value": schema.StringAttribute{
				Optional:    true,
				Description: "Value for CAA record",
			},
			"caa_flags": schema.Int64Attribute{
				Optional:    true,
				Description: "Flags for CAA record",
			},
			"caa_tag": schema.StringAttribute{
				Optional:    true,
				Description: "Tag for CAA record",
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

	// Basic validation: content is required for non-CAA records
	if plan.Type.ValueString() != "CAA" && (plan.Content.IsNull() || plan.Content.ValueString() == "") {
		resp.Diagnostics.AddError("Missing content", "The 'content' attribute is required for non-CAA records.")
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

	// Send CAA fields ONLY for CAA record type to avoid API validation errors
	if plan.Type.ValueString() == "CAA" {
		if !plan.CAAValue.IsNull() && !plan.CAAValue.IsUnknown() {
			createReq.CAAValue = plan.CAAValue.ValueString()
		}
		if !plan.CAAFlags.IsNull() && !plan.CAAFlags.IsUnknown() {
			f := plan.CAAFlags.ValueInt64()
			createReq.Flags = &f
		}
		if !plan.CAATag.IsNull() && !plan.CAATag.IsUnknown() {
			createReq.Tag = plan.CAATag.ValueString()
		}

		// Active24 requires 'content' even for CAA records. Use caaValue as content.
		if plan.Content.IsNull() || plan.Content.ValueString() == "" {
			createReq.Content = plan.CAAValue.ValueString()
		}
	}

	targetService := plan.Domain.ValueString()
	if !plan.Service.IsNull() && !plan.Service.IsUnknown() && plan.Service.ValueString() != "" {
		targetService = plan.Service.ValueString()
	}

	createdRec, err := r.client.CreateRecord(ctx, targetService, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating record", err.Error())
		return
	}

	// Prefer the record returned by Create (it should have the ID)
	rec := createdRec
	if rec == nil || rec.ID == 0 {
		// Fallback: search by name+type, then match by content to find the exact record
		records, err := r.client.ListRecords(ctx, targetService, normalizeNameForAPI(plan.Name.ValueString()), plan.Type.ValueString(), "", nil)
		if err != nil || len(records) == 0 {
			resp.Diagnostics.AddError("Error reading created record", fmt.Sprintf("lookup failed: %v", err))
			return
		}
		// Match by content (or caaValue for CAA) to find the exact record among duplicates
		matchContent := createReq.Content
		for i := range records {
			if records[i].Content == matchContent || records[i].CAAValue == matchContent {
				rec = &records[i]
				break
			}
		}
		if rec == nil {
			// Last resort: take the first match
			rec = &records[0]
		}
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d", rec.ID))
	// Only update TTL from API if it returned a non-zero value
	if rec.TTL > 0 {
		plan.TTL = types.Int64Value(rec.TTL)
	}
	if rec.Priority != nil {
		plan.Priority = types.Int64Value(*rec.Priority)
	} else {
		plan.Priority = types.Int64Null()
	}

	// For CAA: update caa_* fields if API returns them, keep content null
	if strings.EqualFold(plan.Type.ValueString(), "CAA") {
		if rec.CAAValue != "" {
			plan.CAAValue = types.StringValue(rec.CAAValue)
		}
		if rec.Flags != nil {
			plan.CAAFlags = types.Int64Value(*rec.Flags)
		}
		if rec.Tag != "" {
			plan.CAATag = types.StringValue(rec.Tag)
		}
		plan.Content = types.StringNull()
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

	// Try to get record by ID directly
	rec, err := r.client.GetRecord(ctx, targetService, id)
	if err != nil {
		// Fallback: try list records if GetRecord fails (some record types/APIs might behave differently)
		records, _ := r.client.ListRecords(ctx, targetService, normalizeNameForAPI(state.Name.ValueString()), state.Type.ValueString(), "", nil)
		for i := range records {
			if records[i].ID == id {
				rec = &records[i]
				break
			}
		}
	}

	if rec == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// If API returns FQDN, strip the domain part to match relative names in TF config
	apiName := rec.Name
	domain := state.Domain.ValueString()

	if apiName == domain || apiName == domain+"." {
		apiName = "" // will be "@" after denormalizeNameFromAPI
	} else {
		suffix := "." + domain
		if strings.HasSuffix(apiName, suffix) {
			apiName = strings.TrimSuffix(apiName, suffix)
		}
	}

	state.Name = types.StringValue(denormalizeNameFromAPI(apiName))
	state.Type = types.StringValue(rec.Type)
	state.TTL = types.Int64Value(rec.TTL)

	if rec.Priority != nil {
		state.Priority = types.Int64Value(*rec.Priority)
	} else {
		state.Priority = types.Int64Null()
	}

	if strings.EqualFold(rec.Type, "CAA") {
		// For CAA records: populate caa_* fields from API.
		// Do NOT set content in state - it is only used internally for API calls.
		if rec.CAAValue != "" {
			state.CAAValue = types.StringValue(rec.CAAValue)
		} else if rec.Content != "" && (state.CAAValue.IsNull() || state.CAAValue.ValueString() == "") {
			// API didn't return caaValue separately, use content as fallback
			state.CAAValue = types.StringValue(rec.Content)
		}
		if rec.Flags != nil {
			state.CAAFlags = types.Int64Value(*rec.Flags)
		} else if state.CAAFlags.IsNull() {
			state.CAAFlags = types.Int64Value(0)
		}
		if rec.Tag != "" {
			state.CAATag = types.StringValue(rec.Tag)
		}
		// Keep content null for CAA - user should not need to set it
		state.Content = types.StringNull()
	} else {
		// For all other record types: always sync content from API
		state.Content = types.StringValue(rec.Content)
		// Ensure CAA fields are null for non-CAA records
		state.CAAValue = types.StringNull()
		state.CAAFlags = types.Int64Null()
		state.CAATag = types.StringNull()
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

	// Basic validation: content is required for non-CAA records
	if plan.Type.ValueString() != "CAA" && (plan.Content.IsNull() || plan.Content.ValueString() == "") {
		resp.Diagnostics.AddError("Missing content", "The 'content' attribute is required for non-CAA records.")
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

	// Send CAA fields ONLY for CAA record type to avoid API validation errors
	if plan.Type.ValueString() == "CAA" {
		if !plan.CAAValue.IsNull() && !plan.CAAValue.IsUnknown() {
			updateReq.CAAValue = plan.CAAValue.ValueString()
		}
		if !plan.CAAFlags.IsNull() && !plan.CAAFlags.IsUnknown() {
			f := plan.CAAFlags.ValueInt64()
			updateReq.Flags = &f
		}
		if !plan.CAATag.IsNull() && !plan.CAATag.IsUnknown() {
			updateReq.Tag = plan.CAATag.ValueString()
		}

		// Active24 requires 'content' even for CAA records. Use caaValue as content.
		if plan.Content.IsNull() || plan.Content.ValueString() == "" {
			updateReq.Content = plan.CAAValue.ValueString()
		}
	}

	targetService := plan.Domain.ValueString()
	if !plan.Service.IsNull() && !plan.Service.IsUnknown() && plan.Service.ValueString() != "" {
		targetService = plan.Service.ValueString()
	}
	updatedRec, err := r.client.UpdateRecord(ctx, targetService, id, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating record", err.Error())
		return
	}

	// Use updated record if returned, otherwise fetch
	rec := updatedRec
	if rec == nil {
		rec, _ = r.client.GetRecord(ctx, targetService, id)
	}

	if rec != nil {
		// Only update TTL from API if it returned a non-zero value
		if rec.TTL > 0 {
			plan.TTL = types.Int64Value(rec.TTL)
		}
		if rec.Priority != nil {
			plan.Priority = types.Int64Value(*rec.Priority)
		} else {
			plan.Priority = types.Int64Null()
		}
		if rec.CAAValue != "" {
			plan.CAAValue = types.StringValue(rec.CAAValue)
		}
		if rec.Flags != nil {
			plan.CAAFlags = types.Int64Value(*rec.Flags)
		}
		if rec.Tag != "" {
			plan.CAATag = types.StringValue(rec.Tag)
		}
	}

	// For CAA: keep content null in state
	if strings.EqualFold(plan.Type.ValueString(), "CAA") {
		plan.Content = types.StringNull()
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
	// Supported formats:
	//   <domain>:<id>                                e.g. finbricks.com:311059968
	//   <domain>:<service>:<id>                      e.g. finbricks.com:12905048:311059968
	//   <domain>:<name>:<type>                       e.g. finbricks.com:devtest.dev:A
	//   <domain>:<service>:<name>:<type>             e.g. finbricks.com:12905048:devtest.dev:CAA
	//   <domain>:<service>:<name>:<type>:<content>   e.g. finbricks.com:12905048:@:CAA:letsencrypt.org
	//
	// The last format allows disambiguating when multiple records of the same
	// type exist on the same name (e.g. multiple CAA records).
	//
	// The provider auto-detects the format: if the last (or second-to-last)
	// part is a known DNS record type string, it uses name+type lookup;
	// otherwise it treats the last part as a numeric record ID.

	parts := strings.Split(req.ID, ":")

	switch len(parts) {
	case 2:
		// <domain>:<id>
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)

	case 3:
		if isDNSType(parts[2]) {
			// <domain>:<name>:<type>
			r.importByNameType(ctx, parts[0], "", parts[1], parts[2], "", resp)
		} else {
			// <domain>:<service>:<id>
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service"), parts[1])...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
		}

	case 4:
		// <domain>:<service>:<name>:<type>
		r.importByNameType(ctx, parts[0], parts[1], parts[2], parts[3], "", resp)

	case 5:
		// <domain>:<service>:<name>:<type>:<content>
		r.importByNameType(ctx, parts[0], parts[1], parts[2], parts[3], parts[4], resp)

	default:
		resp.Diagnostics.AddError("Invalid import format",
			"Expected one of:\n"+
				"  <domain>:<id>\n"+
				"  <domain>:<service>:<id>\n"+
				"  <domain>:<name>:<type>\n"+
				"  <domain>:<service>:<name>:<type>\n"+
				"  <domain>:<service>:<name>:<type>:<content>")
	}
}

// importByNameType looks up a record by name and type via the API, then sets the state.
// If content is non-empty, it is used to disambiguate when multiple records match name+type.
func (r *dnsRecordResource) importByNameType(ctx context.Context, domain, service, name, rtype, content string, resp *resource.ImportStateResponse) {
	targetService := domain
	if service != "" {
		targetService = service
	}

	// Normalize "@" to empty string for API
	if name == "@" {
		name = ""
	}

	// Strip domain suffix if user provided FQDN (e.g. "devtest.dev.finbricks.com" -> "devtest.dev")
	fqdnSuffix := "." + domain
	if strings.HasSuffix(name, fqdnSuffix) {
		name = strings.TrimSuffix(name, fqdnSuffix)
	} else if name == domain {
		name = "" // apex
	}

	records, err := r.client.ListRecords(ctx, targetService, name, strings.ToUpper(rtype), "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Error looking up record", fmt.Sprintf("API error: %v", err))
		return
	}

	// Find matching records by name and type
	var matches []DNSRecord
	for i := range records {
		recName := records[i].Name
		// Normalize: strip domain suffix from API response too
		if strings.HasSuffix(recName, fqdnSuffix) {
			recName = strings.TrimSuffix(recName, fqdnSuffix)
		}
		if recName == name && strings.EqualFold(records[i].Type, rtype) {
			matches = append(matches, records[i])
		}
	}

	if len(matches) == 0 {
		resp.Diagnostics.AddError("Record not found",
			fmt.Sprintf("No %s record named '%s' found in zone %s (service %s)", rtype, name, domain, targetService))
		return
	}

	var found *DNSRecord

	if len(matches) == 1 {
		found = &matches[0]
	} else {
		// Multiple matches - try to disambiguate by content
		if content != "" {
			for i := range matches {
				recContent := matches[i].Content
				recCAAValue := matches[i].CAAValue
				if strings.EqualFold(recContent, content) || strings.EqualFold(recCAAValue, content) {
					found = &matches[i]
					break
				}
			}
			if found == nil {
				resp.Diagnostics.AddError("Record not found",
					fmt.Sprintf("Found %d %s records named '%s' but none matching content '%s'", len(matches), rtype, name, content))
				return
			}
		} else {
			// No content filter - show all matches so user can pick
			var ids []string
			for _, m := range matches {
				detail := m.Content
				if m.CAAValue != "" {
					detail = fmt.Sprintf("%s (tag=%s)", m.CAAValue, m.Tag)
				}
				ids = append(ids, fmt.Sprintf("  ID %d: %s", m.ID, detail))
			}
			resp.Diagnostics.AddError("Multiple records found",
				fmt.Sprintf("Found %d %s records named '%s'. Use one of these formats to disambiguate:\n"+
					"  <domain>:<service>:<name>:<type>:<content>\n"+
					"  <domain>:<service>:<id>\n\n"+
					"Matching records:\n%s", len(matches), rtype, name, strings.Join(ids, "\n")))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), domain)...)
	if service != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service"), service)...)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), fmt.Sprintf("%d", found.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), denormalizeNameFromAPI(name))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), strings.ToUpper(rtype))...)
}

// isDNSType checks if a string is a known DNS record type.
func isDNSType(s string) bool {
	switch strings.ToUpper(s) {
	case "A", "AAAA", "CNAME", "MX", "TXT", "SRV", "NS", "CAA", "SOA", "PTR", "TLSA", "SSHFP":
		return true
	}
	return false
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
