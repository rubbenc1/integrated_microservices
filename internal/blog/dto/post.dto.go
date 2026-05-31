package dto

type ListPostsParams struct {
	Limit  int
	Offset int
}

type UpdatePostRequest struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
}

type Req struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
