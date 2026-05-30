package registry

import (
	"fmt"
	"time"
)

type Claim struct {
	Project     string    `json:"project"`
	Service     string    `json:"service"`
	Description string    `json:"description"`
	ClaimedAt   time.Time `json:"claimed_at"`
	Path        string    `json:"path"`
}

func (r *Registry) AddClaim(port, project, service, desc, path string, force bool) error {
	existing, ok := r.Get(port)
	if ok {
		if existing.Project == project && existing.Service == service {
			// same owner: just refresh the description and timestamp
			existing.Description = desc
			existing.ClaimedAt = time.Now().UTC()
			r.Set(port, existing)
			return nil
		}
		if !force {
			return fmt.Errorf(
				"port %s is already claimed by %s/%s",
				port,
				existing.Project,
				existing.Service,
			)
		}
	}
	// new claim, or forced overwrite of an existing one
	r.Set(port, Claim{
		Project:     project,
		Service:     service,
		Description: desc,
		ClaimedAt:   time.Now().UTC(),
		Path:        path,
	})
	return nil
}

func (r *Registry) RemoveClaim(port string) {
	r.Delete(port)
}

func (r *Registry) GetClaim(port string) *Claim {
	c, ok := r.Get(port)
	if !ok {
		return nil
	}
	return &c
}

func (r *Registry) FilterByProject(project string) []Claim {
	result := []Claim{}
	for _, c := range r.Claims {
		if project == "" || c.Project == project {
			result = append(result, c)
		}
	}
	return result
}
