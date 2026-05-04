package remoteindex

import "context"

type FakeClient struct {
	StatusResponse  StatusResponse
	RefreshResponse RefreshResponse
	SearchResponse  SearchResponse
	Calls           []string
}

func (c *FakeClient) Status(ctx context.Context, req StatusRequest) (StatusResponse, error) {
	if err := ctx.Err(); err != nil {
		return StatusResponse{}, err
	}
	c.Calls = append(c.Calls, "status:"+req.WorkspaceID)
	return c.StatusResponse, nil
}

func (c *FakeClient) Refresh(ctx context.Context, req RefreshRequest) (RefreshResponse, error) {
	if err := ctx.Err(); err != nil {
		return RefreshResponse{}, err
	}
	c.Calls = append(c.Calls, "refresh:"+req.WorkspaceID)
	return c.RefreshResponse, nil
}

func (c *FakeClient) Search(ctx context.Context, req SearchRequest) (SearchResponse, error) {
	if err := ctx.Err(); err != nil {
		return SearchResponse{}, err
	}
	c.Calls = append(c.Calls, "search:"+req.WorkspaceID+":"+req.Query)
	return c.SearchResponse, nil
}
