package main

// Will unmarshall and keep a cache of some of the embeded json files.
type DataCache struct {
	DailyQuests   []DailyQuest
	Resources     []Resource
	UnitTemplates []UnitTemplate
	UnitTypes     []UnitType
}

func (dc *DataCache) Load() error {
	var err error

	dc.DailyQuests, err = UnmarshallDailyQuestsJson()
	if err != nil {
		return err
	}

	dc.Resources, err = UnmarshallResourcesJson()
	if err != nil {
		return err
	}

	dc.UnitTemplates, err = UnMarshallUnitTemplatesJson()
	if err != nil {
		return err
	}

	dc.UnitTypes, err = UnMarshallUnitTypesJson()
	if err != nil {
		return err
	}

	return nil
}
