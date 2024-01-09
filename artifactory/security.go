package artifactory

import (
	"encoding/json"
)

const (
	usersEndpoint  = "security/users"
	groupsEndpoint = "security/groups"
)

// User represents single element of API respond from users endpoint
type User struct {
	Name  string `json:"name"`
	Realm string `json:"realm"`
}
type Users struct {
	Users  []User
	NodeId string
}

// FetchUsers makes the API call to users endpoint and returns []User
func (c *Client) FetchUsers() (Users, error) {
	var users Users
	c.logger.Debug("Fetching users stats")
	resp, err := c.FetchHTTP(usersEndpoint)
	if err != nil {
		return users, err
	}
	users.NodeId = resp.NodeId
	if err := json.Unmarshal(resp.Body, &users.Users); err != nil {
		c.logger.Error("There was an issue when try to unmarshal users respond")
		return users, &UnmarshalError{
			message:  err.Error(),
			endpoint: usersEndpoint,
		}
	}
	return users, nil
}

// Group represents single element of API respond from groups endpoint
type Group struct {
	Name  string `json:"name"`
	Realm string `json:"uri"`
}
type Groups struct {
	Groups []Group
	NodeId string
}

// FetchGroups makes the API call to groups endpoint and returns []Group
func (c *Client) FetchGroups() (Groups, error) {
	var groups Groups
	c.logger.Debug("Fetching groups stats")
	resp, err := c.FetchHTTP(groupsEndpoint)
	if err != nil {
		return groups, err
	}
	groups.NodeId = resp.NodeId
	if err := json.Unmarshal(resp.Body, &groups.Groups); err != nil {
		c.logger.Error("There was an issue when try to unmarshal groups respond")
		return groups, &UnmarshalError{
			message:  err.Error(),
			endpoint: groupsEndpoint,
		}
	}

	return groups, nil
}
