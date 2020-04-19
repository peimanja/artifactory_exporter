package artifactory

import (
	"encoding/json"

	"github.com/go-kit/kit/log/level"
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

// FetchUsers makes the API call to users endpoint and returns []User
func (c *Client) FetchUsers() ([]User, error) {
	var users []User
	level.Debug(c.logger).Log("msg", "Fetching users stats")
	resp, err := c.fetchHTTP(usersEndpoint)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resp, &users); err != nil {
		level.Error(c.logger).Log("msg", "There was an issue when try to unmarshal users respond")
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

// FetchGroups makes the API call to groups endpoint and returns []Group
func (c *Client) FetchGroups() ([]Group, error) {
	var groups []Group
	level.Debug(c.logger).Log("msg", "Fetching groups stats")
	resp, err := c.fetchHTTP(groupsEndpoint)
	if err != nil {
		return groups, err
	}
	if err := json.Unmarshal(resp, &groups); err != nil {
		level.Error(c.logger).Log("msg", "There was an issue when try to unmarshal groups respond")
		return groups, &UnmarshalError{
			message:  err.Error(),
			endpoint: groupsEndpoint,
		}
	}

	return groups, nil
}
