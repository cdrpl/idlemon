package main

// Will unmarshall and keep a cache of some of the embeded json files.
type DataCache struct {
	DailyQuests   []DailyQuest
	Resources     []Resource
	UnitTemplates []UnitTemplate
}

func (dc *DataCache) Load() error {
	var err error

	dc.DailyQuests = DailyQuests()
	dc.Resources = Resources()
	dc.UnitTemplates, err = UnMarshallUnitTemplatesJson()

	return err
}
