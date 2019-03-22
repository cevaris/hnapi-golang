package model

// Items is for serializin json
type Items struct {
	Items    []Item `json:"items"`
	Comments []Item `json:"comments,omitempty"`
}

// Item is either Story, Comment, or Poll
type Item struct {
	ID   int    `json:"id"`
	Type string `json:"type,omitempty"`
	By   string `json:"by,omitempty"`
	Time int    `json:"time,omitempty"`

	Deleted bool `json:"deleted,omitempty"`
	Dead    bool `json:"dead,omitempty"`

	Parent int `json:"parent,omitempty"`

	Poll  int   `json:"poll,omitempty"`
	Parts []int `json:"parts,omitempty"`

	Decendants int   `json:"decendants,omitempty"`
	Kids       []int `json:"kids,omitempty"`

	URL   string `json:"url,omitempty"`
	Score int    `json:"score,omitempty"`
	Title string `json:"title,omitempty"`
}
