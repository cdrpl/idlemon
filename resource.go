package main

// Represents a monetary resource such as gold.
type Resource struct {
	ID     int `json:"id"`
	Amount int `json:"amount"`
}

// Return a slice of all resources in the game.
func Resources() []Resource {
	resources := make([]Resource, 4)

	resources[RESOURCE_GOLD] = Resource{ID: RESOURCE_GOLD}
	resources[RESOURCE_GEMS] = Resource{ID: RESOURCE_GEMS}
	resources[RESOURCE_EXP_STONE] = Resource{ID: RESOURCE_EXP_STONE}
	resources[RESOURCE_EVO_STONE] = Resource{ID: RESOURCE_EVO_STONE}

	return resources
}
