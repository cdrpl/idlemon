package main

// Will keep a cache of game data that doesn't get stored in the database.
type DataCache struct {
	DailyQuests   []DailyQuest
	Resources     []Resource
	UnitTemplates []UnitTemplate
}

// Initialize the DataCache variables.
func (dc *DataCache) Load() error {
	var err error

	dc.DailyQuests = DailyQuests()
	dc.Resources = Resources()
	dc.UnitTemplates, err = UnMarshalUnitTemplatesJson()

	return err
}
