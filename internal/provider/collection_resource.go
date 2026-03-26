package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &CollectionResource{}
	_ resource.ResourceWithConfigure = &CollectionResource{}
)

type CollectionResource struct {
	client *Client
}

type CollectionResourceModel struct {
	ID              types.String            `tfsdk:"id"`
	Title           types.String            `tfsdk:"title"`
	Handle          types.String            `tfsdk:"handle"`
	DescriptionHTML types.String            `tfsdk:"description_html"`
	SortOrder       types.String            `tfsdk:"sort_order"`
	TemplateSuffix  types.String            `tfsdk:"template_suffix"`
	RuleSet         *CollectionRuleSetModel `tfsdk:"rule_set"`
}

type CollectionRuleSetModel struct {
	AppliedDisjunctively types.Bool            `tfsdk:"applied_disjunctively"`
	Rules                []CollectionRuleModel `tfsdk:"rules"`
}

type CollectionRuleModel struct {
	Column    types.String `tfsdk:"column"`
	Condition types.String `tfsdk:"condition"`
	Relation  types.String `tfsdk:"relation"`
}

// API response types.

type collectionCreateData struct {
	CollectionCreate struct {
		Collection *collectionData `json:"collection"`
		UserErrors []userError     `json:"userErrors"`
	} `json:"collectionCreate"`
}

type collectionUpdateData struct {
	CollectionUpdate struct {
		Collection *collectionData `json:"collection"`
		UserErrors []userError     `json:"userErrors"`
	} `json:"collectionUpdate"`
}

type collectionDeleteData struct {
	CollectionDelete struct {
		DeletedCollectionID string      `json:"deletedCollectionId"`
		UserErrors          []userError `json:"userErrors"`
	} `json:"collectionDelete"`
}

type collectionReadData struct {
	Collection *collectionData `json:"collection"`
}

type collectionData struct {
	ID              string       `json:"id"`
	Title           string       `json:"title"`
	Handle          string       `json:"handle"`
	DescriptionHTML string       `json:"descriptionHtml"`
	SortOrder       string       `json:"sortOrder"`
	TemplateSuffix  *string      `json:"templateSuffix"`
	RuleSet         *ruleSetData `json:"ruleSet"`
}

type ruleSetData struct {
	AppliedDisjunctively bool       `json:"appliedDisjunctively"`
	Rules                []ruleData `json:"rules"`
}

type ruleData struct {
	Column    string `json:"column"`
	Condition string `json:"condition"`
	Relation  string `json:"relation"`
}

func NewCollectionResource() resource.Resource {
	return &CollectionResource{}
}

func (r *CollectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collection"
}

func (r *CollectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shopify collection (manual or smart).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Shopify collection GID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Description: "Collection title.",
				Required:    true,
			},
			"handle": schema.StringAttribute{
				Description: "URL-friendly collection handle. Auto-generated from title if not set.",
				Optional:    true,
				Computed:    true,
			},
			"description_html": schema.StringAttribute{
				Description: "Collection description in HTML.",
				Optional:    true,
				Computed:    true,
			},
			"sort_order": schema.StringAttribute{
				Description: "Product sort order: ALPHA_ASC, ALPHA_DESC, BEST_SELLING, CREATED, CREATED_DESC, MANUAL, PRICE_ASC, PRICE_DESC.",
				Optional:    true,
				Computed:    true,
			},
			"template_suffix": schema.StringAttribute{
				Description: "Liquid template suffix for the collection page.",
				Optional:    true,
				Computed:    true,
			},
			"rule_set": schema.SingleNestedAttribute{
				Description: "Rules for automatic (smart) collections. Omit for manual collections.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"applied_disjunctively": schema.BoolAttribute{
						Description: "Whether products must match any rule (true) or all rules (false).",
						Required:    true,
					},
					"rules": schema.ListNestedAttribute{
						Description: "Collection rules.",
						Required:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"column": schema.StringAttribute{
									Description: "Rule column: TAG, TITLE, TYPE, VENDOR, VARIANT_PRICE, IS_PRICE_REDUCED, VARIANT_COMPARE_AT_PRICE, VARIANT_WEIGHT, VARIANT_INVENTORY, VARIANT_TITLE.",
									Required:    true,
								},
								"condition": schema.StringAttribute{
									Description: "Rule condition value.",
									Required:    true,
								},
								"relation": schema.StringAttribute{
									Description: "Rule relation: EQUALS, NOT_EQUALS, CONTAINS, NOT_CONTAINS, STARTS_WITH, ENDS_WITH, GREATER_THAN, LESS_THAN, IS_SET, IS_NOT_SET.",
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *CollectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *CollectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CollectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]any{
		"title": plan.Title.ValueString(),
	}
	if !plan.Handle.IsNull() && !plan.Handle.IsUnknown() {
		input["handle"] = plan.Handle.ValueString()
	}
	if !plan.DescriptionHTML.IsNull() && !plan.DescriptionHTML.IsUnknown() {
		input["descriptionHtml"] = plan.DescriptionHTML.ValueString()
	}
	if !plan.SortOrder.IsNull() && !plan.SortOrder.IsUnknown() {
		input["sortOrder"] = plan.SortOrder.ValueString()
	}
	if !plan.TemplateSuffix.IsNull() && !plan.TemplateSuffix.IsUnknown() {
		input["templateSuffix"] = plan.TemplateSuffix.ValueString()
	}
	if plan.RuleSet != nil {
		input["ruleSet"] = buildRuleSetInput(plan.RuleSet)
	}

	result, err := r.client.Execute(queryCollectionCreate, map[string]any{"input": input})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create collection", err.Error())
		return
	}

	var data collectionCreateData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}
	if len(data.CollectionCreate.UserErrors) > 0 {
		resp.Diagnostics.AddError("Shopify user error", formatUserErrors(data.CollectionCreate.UserErrors))
		return
	}
	if data.CollectionCreate.Collection == nil {
		resp.Diagnostics.AddError("Shopify returned nil collection", "collectionCreate returned null collection without errors")
		return
	}

	mapCollectionToState(data.CollectionCreate.Collection, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CollectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CollectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Execute(queryCollectionRead, map[string]any{"id": state.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read collection", err.Error())
		return
	}

	var data collectionReadData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	if data.Collection == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	mapCollectionToState(data.Collection, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CollectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CollectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CollectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := map[string]any{
		"id":    state.ID.ValueString(),
		"title": plan.Title.ValueString(),
	}
	if !plan.Handle.IsNull() && !plan.Handle.IsUnknown() {
		input["handle"] = plan.Handle.ValueString()
	}
	if !plan.DescriptionHTML.IsNull() && !plan.DescriptionHTML.IsUnknown() {
		input["descriptionHtml"] = plan.DescriptionHTML.ValueString()
	}
	if !plan.SortOrder.IsNull() && !plan.SortOrder.IsUnknown() {
		input["sortOrder"] = plan.SortOrder.ValueString()
	}
	if !plan.TemplateSuffix.IsNull() && !plan.TemplateSuffix.IsUnknown() {
		input["templateSuffix"] = plan.TemplateSuffix.ValueString()
	}
	if plan.RuleSet != nil {
		input["ruleSet"] = buildRuleSetInput(plan.RuleSet)
	}

	result, err := r.client.Execute(queryCollectionUpdate, map[string]any{"input": input})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update collection", err.Error())
		return
	}

	var data collectionUpdateData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		resp.Diagnostics.AddError("Failed to parse update response", err.Error())
		return
	}
	if len(data.CollectionUpdate.UserErrors) > 0 {
		resp.Diagnostics.AddError("Shopify user error", formatUserErrors(data.CollectionUpdate.UserErrors))
		return
	}

	mapCollectionToState(data.CollectionUpdate.Collection, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CollectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CollectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	variables := map[string]any{
		"input": map[string]any{
			"id": state.ID.ValueString(),
		},
	}

	result, err := r.client.Execute(queryCollectionDelete, variables)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete collection", err.Error())
		return
	}

	var data collectionDeleteData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		resp.Diagnostics.AddError("Failed to parse delete response", err.Error())
		return
	}
	if len(data.CollectionDelete.UserErrors) > 0 {
		resp.Diagnostics.AddError("Shopify user error", formatUserErrors(data.CollectionDelete.UserErrors))
		return
	}
}

// buildRuleSetInput converts the Terraform model to a GraphQL ruleSet input.
func buildRuleSetInput(rs *CollectionRuleSetModel) map[string]any {
	rules := make([]map[string]any, len(rs.Rules))
	for i, rule := range rs.Rules {
		rules[i] = map[string]any{
			"column":    rule.Column.ValueString(),
			"condition": rule.Condition.ValueString(),
			"relation":  rule.Relation.ValueString(),
		}
	}
	return map[string]any{
		"appliedDisjunctively": rs.AppliedDisjunctively.ValueBool(),
		"rules":                rules,
	}
}

// mapCollectionToState updates the Terraform state from a Shopify collection response.
func mapCollectionToState(c *collectionData, model *CollectionResourceModel) {
	model.ID = types.StringValue(c.ID)
	model.Title = types.StringValue(c.Title)
	model.Handle = types.StringValue(c.Handle)
	model.DescriptionHTML = types.StringValue(c.DescriptionHTML)
	model.SortOrder = types.StringValue(c.SortOrder)
	model.TemplateSuffix = nullableString(c.TemplateSuffix)

	if c.RuleSet != nil {
		rules := make([]CollectionRuleModel, len(c.RuleSet.Rules))
		for i, rule := range c.RuleSet.Rules {
			rules[i] = CollectionRuleModel{
				Column:    types.StringValue(rule.Column),
				Condition: types.StringValue(rule.Condition),
				Relation:  types.StringValue(rule.Relation),
			}
		}
		model.RuleSet = &CollectionRuleSetModel{
			AppliedDisjunctively: types.BoolValue(c.RuleSet.AppliedDisjunctively),
			Rules:                rules,
		}
	} else {
		model.RuleSet = nil
	}
}
