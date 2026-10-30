package gorest

// Resource contains the common fields in Asana's compact resource objects.
type Resource struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
}

// User contains the fields returned by Asana's compact user representation.
type User struct {
	Resource
	Name string `json:"name"`
}

// Project contains the fields returned by Asana's compact project
// representation.
type Project struct {
	Resource
	Name            string `json:"name"`
	ResourceSubtype string `json:"resource_subtype,omitempty"`
}

// NextPage contains the cursor returned by Asana for the next page.
type NextPage struct {
	Offset string `json:"offset,omitempty"`
}

type UsersResponse struct {
	Data     []User    `json:"data"`
	NextPage *NextPage `json:"next_page,omitempty"`
}

type ProjectsResponse struct {
	Data     []Project `json:"data"`
	NextPage *NextPage `json:"next_page,omitempty"`
}
