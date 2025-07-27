package provider

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "io"

   // "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UserResource{}

func NewUserResource() resource.Resource {
    return &UserResource{}
}

type UserResource struct {
    client *APIClient
}

type UserResourceModel struct {
    ID       types.Int64  `tfsdk:"id"`
    Name     types.String `tfsdk:"name"`
    Email    types.String `tfsdk:"email"`
    Username types.String `tfsdk:"username"`
}

// -------------------------
// Metadata
// -------------------------
func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_user"
}

// -------------------------
// Schema
// -------------------------
func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "User resource",
        Attributes: map[string]schema.Attribute{
            "id": schema.Int64Attribute{
                MarkdownDescription: "User ID",
                Computed:            true,
            },
            "name": schema.StringAttribute{
                MarkdownDescription: "User name",
                Required:            true,
            },
            "email": schema.StringAttribute{
                MarkdownDescription: "User email",
                Required:            true,
            },
            "username": schema.StringAttribute{
                MarkdownDescription: "Username",
                Required:            true,
            },
        },
    }
}

// -------------------------
// Configure
// -------------------------
func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*APIClient)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }

    r.client = client
}

// -------------------------
// Create
// -------------------------
func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data UserResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    user := struct {
        Name     string `json:"name"`
        Email    string `json:"email"`
        Username string `json:"username"`
    }{
        Name:     data.Name.ValueString(),
        Email:    data.Email.ValueString(),
        Username: data.Username.ValueString(),
    }

    body, err := json.Marshal(user)
    if err != nil {
        resp.Diagnostics.AddError("Serialization Error", fmt.Sprintf("Unable to serialize user, got error: %s", err))
        return
    }

    httpResp, err := http.Post(fmt.Sprintf("%s/users", r.client.Endpoint), "application/json", bytes.NewBuffer(body))
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusCreated {
        resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to create user, got status code: %d", httpResp.StatusCode))
        return
    }

    var response struct {
        Message string `json:"message"`
        User    struct {
            ID       int64  `json:"id"`
            Name     string `json:"name"`
            Email    string `json:"email"`
            Username string `json:"username"`
        } `json:"user"`
    }

    err = json.NewDecoder(httpResp.Body).Decode(&response)
    if err != nil {
        resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse user response, got error: %s", err))
        return
    }

    data.ID = types.Int64Value(response.User.ID)
    data.Name = types.StringValue(response.User.Name)
    data.Email = types.StringValue(response.User.Email)
    data.Username = types.StringValue(response.User.Username)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -------------------------
// Read
// -------------------------
func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var data UserResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    httpResp, err := http.Get(fmt.Sprintf("%s/users/%d", r.client.Endpoint, data.ID.ValueInt64()))
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode == http.StatusNotFound {
        resp.State.RemoveResource(ctx)
        return
    }

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read user, got status code: %d", httpResp.StatusCode))
        return
    }

    var user struct {
        ID       int64  `json:"id"`
        Name     string `json:"name"`
        Email    string `json:"email"`
        Username string `json:"username"`
    }

    err = json.NewDecoder(httpResp.Body).Decode(&user)
    if err != nil {
        resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse user response, got error: %s", err))
        return
    }

    data.ID = types.Int64Value(user.ID)
    data.Name = types.StringValue(user.Name)
    data.Email = types.StringValue(user.Email)
    data.Username = types.StringValue(user.Username)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -------------------------
// Update
// -------------------------
// Update implements resource.ResourceWithConfigure
func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan UserResourceModel
    var state UserResourceModel

    // Get planned changes
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Get current state (to fetch ID)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    userID := state.ID.ValueInt64()
    if userID == 0 {
        resp.Diagnostics.AddError("Invalid ID", "User ID is zero. Cannot update a resource with an invalid ID.")
        return
    }

    // Construct request payload
    user := struct {
        Name     string `json:"name"`
        Email    string `json:"email"`
        Username string `json:"username"`
    }{
        Name:     plan.Name.ValueString(),
        Email:    plan.Email.ValueString(),
        Username: plan.Username.ValueString(),
    }

    body, err := json.Marshal(user)
    if err != nil {
        resp.Diagnostics.AddError("Serialization Error", fmt.Sprintf("Unable to serialize user, got error: %s", err))
        return
    }

    url := fmt.Sprintf("%s/users/%d", r.client.Endpoint, userID)

    // Debug logging
    fmt.Printf("[DEBUG] PUT URL: %s\n", url)
    fmt.Printf("[DEBUG] PUT Body: %s\n", string(body))

    // Create PUT request
    httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
        return
    }
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Accept", "application/json")

    httpResp, err := http.DefaultClient.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to send update request, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    respBody, _ := io.ReadAll(httpResp.Body)
    fmt.Printf("[DEBUG] PUT Response Code: %d\n", httpResp.StatusCode)
    fmt.Printf("[DEBUG] PUT Response Body: %s\n", string(respBody))

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update user, got status code: %d, body: %s", httpResp.StatusCode, string(respBody)))
        return
    }

    // Parse updated user from response
    var response struct {
        Message string `json:"message"`
        User    struct {
            ID       int64  `json:"id"`
            Name     string `json:"name"`
            Email    string `json:"email"`
            Username string `json:"username"`
        } `json:"user"`
    }

    err = json.Unmarshal(respBody, &response)
    if err != nil {
        resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse user response, got error: %s", err))
        return
    }

    // Save updated state
    updated := UserResourceModel{
        ID:       types.Int64Value(response.User.ID),
        Name:     types.StringValue(response.User.Name),
        Email:    types.StringValue(response.User.Email),
        Username: types.StringValue(response.User.Username),
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)

    r.Read(ctx, resource.ReadRequest{State: resp.State}, &resource.ReadResponse{
    State:       resp.State,
    Diagnostics: resp.Diagnostics,
})
}



// -------------------------
// Delete
// -------------------------
func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var data UserResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/users/%d", r.client.Endpoint, data.ID.ValueInt64()), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
        return
    }

    httpResp, err := http.DefaultClient.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete user, got status code: %d", httpResp.StatusCode))
        return
    }

    // ✅ This is required
    resp.State.RemoveResource(ctx)
}


// -------------------------
// Import State
// -------------------------
func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    userID, err := strconv.ParseInt(req.ID, 10, 64)
    if err != nil {
        resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Invalid ID format: %s", err))
        return
    }

    // Fetch full user from API
    httpResp, err := http.Get(fmt.Sprintf("%s/users/%d", r.client.Endpoint, userID))
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read user, got status code: %d", httpResp.StatusCode))
        return
    }

    var user struct {
        ID       int64  `json:"id"`
        Name     string `json:"name"`
        Email    string `json:"email"`
        Username string `json:"username"`
    }

    err = json.NewDecoder(httpResp.Body).Decode(&user)
    if err != nil {
        resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse user response, got error: %s", err))
        return
    }

    // Set full state
    resp.Diagnostics.Append(resp.State.Set(ctx, &UserResourceModel{
        ID:       types.Int64Value(user.ID),
        Name:     types.StringValue(user.Name),
        Email:    types.StringValue(user.Email),
        Username: types.StringValue(user.Username),
    })...)
}
