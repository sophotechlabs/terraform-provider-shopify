package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &ProductResource{}
	_ resource.ResourceWithConfigure = &ProductResource{}
)

type ProductResource struct {
	client *Client
}

type ProductResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Title            types.String `tfsdk:"title"`
	Handle           types.String `tfsdk:"handle"`
	Status           types.String `tfsdk:"status"`
	Vendor           types.String `tfsdk:"vendor"`
	ProductType      types.String `tfsdk:"product_type"`
	DescriptionHTML  types.String `tfsdk:"description_html"`
	Tags             types.List   `tfsdk:"tags"`
	DefaultVariantID types.String `tfsdk:"default_variant_id"`
	Price            types.String `tfsdk:"price"`
	CompareAtPrice   types.String `tfsdk:"compare_at_price"`
	SKU              types.String `tfsdk:"sku"`
	Barcode          types.String `tfsdk:"barcode"`
}

// API response types.

type productCreateData struct {
	ProductCreate struct {
		Product    *productData `json:"product"`
		UserErrors []userError  `json:"userErrors"`
	} `json:"productCreate"`
}

type productUpdateData struct {
	ProductUpdate struct {
		Product    *productData `json:"product"`
		UserErrors []userError  `json:"userErrors"`
	} `json:"productUpdate"`
}

type productDeleteData struct {
	ProductDelete struct {
		DeletedProductID string      `json:"deletedProductId"`
		UserErrors       []userError `json:"userErrors"`
	} `json:"productDelete"`
}

type variantsBulkUpdateData struct {
	ProductVariantsBulkUpdate struct {
		ProductVariants []variantData `json:"productVariants"`
		UserErrors      []userError   `json:"userErrors"`
	} `json:"productVariantsBulkUpdate"`
}

type productReadData struct {
	Product *productData `json:"product"`
}

type productData struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Handle          string   `json:"handle"`
	Status          string   `json:"status"`
	Vendor          string   `json:"vendor"`
	ProductType     string   `json:"productType"`
	DescriptionHTML string   `json:"descriptionHtml"`
	Tags            []string `json:"tags"`
	Variants        struct {
		Edges []struct {
			Node variantData `json:"node"`
		} `json:"edges"`
	} `json:"variants"`
}

type variantData struct {
	ID             string  `json:"id"`
	Price          string  `json:"price"`
	CompareAtPrice *string `json:"compareAtPrice"`
	Barcode        *string `json:"barcode"`
	SKU            *string `json:"sku"`
}

func NewProductResource() resource.Resource {
	return &ProductResource{}
}

func (r *ProductResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_product"
}

func (r *ProductResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shopify product and its default variant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Shopify product GID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Description: "Product title.",
				Required:    true,
			},
			"handle": schema.StringAttribute{
				Description: "URL-friendly product handle. Auto-generated from title if not set.",
				Optional:    true,
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Product status: ACTIVE, DRAFT, or ARCHIVED.",
				Required:    true,
			},
			"vendor": schema.StringAttribute{
				Description: "Product vendor.",
				Optional:    true,
				Computed:    true,
			},
			"product_type": schema.StringAttribute{
				Description: "Product type.",
				Optional:    true,
				Computed:    true,
			},
			"description_html": schema.StringAttribute{
				Description: "Product description in HTML.",
				Optional:    true,
				Computed:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Product tags.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"default_variant_id": schema.StringAttribute{
				Description: "GID of the default variant created with the product.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"price": schema.StringAttribute{
				Description: "Default variant price.",
				Optional:    true,
				Computed:    true,
			},
			"compare_at_price": schema.StringAttribute{
				Description: "Default variant compare-at price.",
				Optional:    true,
				Computed:    true,
			},
			"sku": schema.StringAttribute{
				Description: "Default variant SKU.",
				Optional:    true,
				Computed:    true,
			},
			"barcode": schema.StringAttribute{
				Description: "Default variant barcode (EAN-13).",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *ProductResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProductResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProductResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build product input.
	input := map[string]any{
		"title":  plan.Title.ValueString(),
		"status": plan.Status.ValueString(),
	}
	if !plan.Handle.IsNull() && !plan.Handle.IsUnknown() {
		input["handle"] = plan.Handle.ValueString()
	}
	if !plan.Vendor.IsNull() && !plan.Vendor.IsUnknown() {
		input["vendor"] = plan.Vendor.ValueString()
	}
	if !plan.ProductType.IsNull() && !plan.ProductType.IsUnknown() {
		input["productType"] = plan.ProductType.ValueString()
	}
	if !plan.DescriptionHTML.IsNull() && !plan.DescriptionHTML.IsUnknown() {
		input["descriptionHtml"] = plan.DescriptionHTML.ValueString()
	}
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["tags"] = tags
	}

	// Create product.
	result, err := r.client.Execute(queryProductCreate, map[string]any{"input": input})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create product", err.Error())
		return
	}

	var data productCreateData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}
	if len(data.ProductCreate.UserErrors) > 0 {
		resp.Diagnostics.AddError("Shopify user error", formatUserErrors(data.ProductCreate.UserErrors))
		return
	}
	if data.ProductCreate.Product == nil {
		resp.Diagnostics.AddError("Shopify returned nil product", "productCreate returned null product without errors")
		return
	}

	product := data.ProductCreate.Product
	plan.ID = types.StringValue(product.ID)
	plan.Handle = types.StringValue(product.Handle)
	plan.Status = types.StringValue(product.Status)
	plan.Vendor = types.StringValue(product.Vendor)
	plan.ProductType = types.StringValue(product.ProductType)
	plan.DescriptionHTML = types.StringValue(product.DescriptionHTML)

	// Extract default variant ID.
	if len(product.Variants.Edges) > 0 {
		variant := product.Variants.Edges[0].Node
		plan.DefaultVariantID = types.StringValue(variant.ID)

		// Update variant if any variant fields are configured.
		if hasVariantFields(plan) {
			r.updateVariant(&plan, product.ID, variant.ID, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			plan.Price = types.StringValue(variant.Price)
			plan.CompareAtPrice = nullableString(variant.CompareAtPrice)
			plan.SKU = nullableString(variant.SKU)
			plan.Barcode = nullableString(variant.Barcode)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProductResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProductResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Execute(queryProductRead, map[string]any{"id": state.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read product", err.Error())
		return
	}

	var data productReadData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	// Product deleted outside Terraform.
	if data.Product == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	product := data.Product
	state.ID = types.StringValue(product.ID)
	state.Title = types.StringValue(product.Title)
	state.Handle = types.StringValue(product.Handle)
	state.Status = types.StringValue(product.Status)
	state.Vendor = types.StringValue(product.Vendor)
	state.ProductType = types.StringValue(product.ProductType)
	state.DescriptionHTML = types.StringValue(product.DescriptionHTML)

	if len(product.Tags) > 0 {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, product.Tags)
		resp.Diagnostics.Append(diags...)
		state.Tags = tagsList
	} else if !state.Tags.IsNull() {
		// Preserve null if was null, otherwise set empty.
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, []string{})
		resp.Diagnostics.Append(diags...)
		state.Tags = tagsList
	}

	if len(product.Variants.Edges) > 0 {
		variant := product.Variants.Edges[0].Node
		state.DefaultVariantID = types.StringValue(variant.ID)
		state.Price = types.StringValue(variant.Price)
		state.CompareAtPrice = nullableString(variant.CompareAtPrice)
		state.SKU = nullableString(variant.SKU)
		state.Barcode = nullableString(variant.Barcode)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProductResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProductResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ProductResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update product fields.
	input := map[string]any{
		"id":     state.ID.ValueString(),
		"title":  plan.Title.ValueString(),
		"status": plan.Status.ValueString(),
	}
	if !plan.Handle.IsNull() && !plan.Handle.IsUnknown() {
		input["handle"] = plan.Handle.ValueString()
	}
	if !plan.Vendor.IsNull() && !plan.Vendor.IsUnknown() {
		input["vendor"] = plan.Vendor.ValueString()
	}
	if !plan.ProductType.IsNull() && !plan.ProductType.IsUnknown() {
		input["productType"] = plan.ProductType.ValueString()
	}
	if !plan.DescriptionHTML.IsNull() && !plan.DescriptionHTML.IsUnknown() {
		input["descriptionHtml"] = plan.DescriptionHTML.ValueString()
	}
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input["tags"] = tags
	}

	result, err := r.client.Execute(queryProductUpdate, map[string]any{"input": input})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update product", err.Error())
		return
	}

	var data productUpdateData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		resp.Diagnostics.AddError("Failed to parse update response", err.Error())
		return
	}
	if len(data.ProductUpdate.UserErrors) > 0 {
		resp.Diagnostics.AddError("Shopify user error", formatUserErrors(data.ProductUpdate.UserErrors))
		return
	}

	product := data.ProductUpdate.Product
	plan.ID = state.ID
	plan.DefaultVariantID = state.DefaultVariantID
	plan.Handle = types.StringValue(product.Handle)
	plan.Status = types.StringValue(product.Status)
	plan.Vendor = types.StringValue(product.Vendor)
	plan.ProductType = types.StringValue(product.ProductType)
	plan.DescriptionHTML = types.StringValue(product.DescriptionHTML)

	// Update variant if variant fields changed.
	if hasVariantFields(plan) {
		r.updateVariant(&plan, state.ID.ValueString(), state.DefaultVariantID.ValueString(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.Price = state.Price
		plan.CompareAtPrice = state.CompareAtPrice
		plan.SKU = state.SKU
		plan.Barcode = state.Barcode
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProductResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProductResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	variables := map[string]any{
		"input": map[string]any{
			"id": state.ID.ValueString(),
		},
	}

	result, err := r.client.Execute(queryProductDelete, variables)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete product", err.Error())
		return
	}

	var data productDeleteData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		resp.Diagnostics.AddError("Failed to parse delete response", err.Error())
		return
	}
	if len(data.ProductDelete.UserErrors) > 0 {
		resp.Diagnostics.AddError("Shopify user error", formatUserErrors(data.ProductDelete.UserErrors))
		return
	}
}

// updateVariant calls productVariantsBulkUpdate and writes variant state.
func (r *ProductResource) updateVariant(plan *ProductResourceModel, productID, variantID string, diagnostics *diag.Diagnostics) {
	variantInput := map[string]any{"id": variantID}

	if !plan.Price.IsNull() && !plan.Price.IsUnknown() {
		variantInput["price"] = plan.Price.ValueString()
	}
	if !plan.CompareAtPrice.IsNull() && !plan.CompareAtPrice.IsUnknown() {
		variantInput["compareAtPrice"] = plan.CompareAtPrice.ValueString()
	}
	if !plan.SKU.IsNull() && !plan.SKU.IsUnknown() {
		variantInput["sku"] = plan.SKU.ValueString()
	}
	if !plan.Barcode.IsNull() && !plan.Barcode.IsUnknown() {
		variantInput["barcode"] = plan.Barcode.ValueString()
	}

	variables := map[string]any{
		"productId": productID,
		"variants":  []any{variantInput},
	}

	result, err := r.client.Execute(queryProductVariantsBulkUpdate, variables)
	if err != nil {
		diagnostics.AddError("Failed to update variant", err.Error())
		return
	}

	var data variantsBulkUpdateData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		diagnostics.AddError("Failed to parse variant update response", err.Error())
		return
	}
	if len(data.ProductVariantsBulkUpdate.UserErrors) > 0 {
		diagnostics.AddError("Shopify variant error", formatUserErrors(data.ProductVariantsBulkUpdate.UserErrors))
		return
	}

	if len(data.ProductVariantsBulkUpdate.ProductVariants) > 0 {
		v := data.ProductVariantsBulkUpdate.ProductVariants[0]
		plan.Price = types.StringValue(v.Price)
		plan.CompareAtPrice = nullableString(v.CompareAtPrice)
		plan.SKU = nullableString(v.SKU)
		plan.Barcode = nullableString(v.Barcode)
	}
}

// hasVariantFields returns true if any variant field is explicitly configured.
func hasVariantFields(plan ProductResourceModel) bool {
	return (!plan.Price.IsNull() && !plan.Price.IsUnknown()) ||
		(!plan.CompareAtPrice.IsNull() && !plan.CompareAtPrice.IsUnknown()) ||
		(!plan.SKU.IsNull() && !plan.SKU.IsUnknown()) ||
		(!plan.Barcode.IsNull() && !plan.Barcode.IsUnknown())
}

// nullableString converts a *string to types.String, returning null for nil.
func nullableString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// formatUserErrors joins Shopify user errors into a single message.
func formatUserErrors(errors []userError) string {
	msg := ""
	for i, e := range errors {
		if i > 0 {
			msg += "; "
		}
		msg += e.Message
	}
	return msg
}
