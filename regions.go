package policefeed

import "fmt"

type Regions struct {
	baseRSSURL string
	m          map[string]Region
	regionIDs  []string
}

func NewRegions(baseRSSURL string) Regions {
	var r Regions
	r.baseRSSURL = baseRSSURL
	r.m = map[string]Region{
		"blekinge":        {ID: "blekinge", Name: "Blekinge"},
		"dalarna":         {ID: "dalarna", Name: "Dalarna"},
		"gotland":         {ID: "gotland", Name: "Gotland"},
		"gavleborg":       {ID: "gavleborg", Name: "Gävleborg"},
		"halland":         {ID: "halland", Name: "Halland"},
		"jamtland":        {ID: "jamtland", Name: "Jämtland"},
		"jonkoping":       {ID: "jonkoping", Name: "Jönköping"},
		"kalmar-lan":      {ID: "kalmar-lan", Name: "Kalmar Län"},
		"kronoberg":       {ID: "kronoberg", Name: "Kronoberg"},
		"norrbotten":      {ID: "norrbotten", Name: "Norrbotten"},
		"skane":           {ID: "skane", Name: "Skåne"},
		"sodermanland":    {ID: "sodermanland", Name: "Södermanland"},
		"stockholms-lan":  {ID: "stockholms-lan", Name: "Stockholms Län"},
		"uppsala-lan":     {ID: "uppsala-lan", Name: "Uppsala Län"},
		"varmland":        {ID: "varmland", Name: "Värmland"},
		"vasterbotten":    {ID: "vasterbotten", Name: "Västerbotten"},
		"vasternorrland":  {ID: "vasternorrland", Name: "Västernorrland"},
		"vastmanland":     {ID: "vastmanland", Name: "Västmanland"},
		"vastra-gotaland": {ID: "vastra-gotaland", Name: "Västra Götaland"},
		"orebro-lan":      {ID: "orebro-lan", Name: "Örebro Län"},
		"ostergotland":    {ID: "ostergotland", Name: "Östergötland"},
	}
	r.regionIDs = make([]string, 0, len(r.m))
	for k := range r.m {
		r.regionIDs = append(r.regionIDs, k)
	}
	return r
}

func (r *Regions) Exists(regionID string) bool {
	_, exists := r.m[regionID]
	return exists
}

func (r *Regions) ListIDs() []string {
	return r.regionIDs
}

func (r *Regions) GetRSSURL(regionID string) string {
	if regionID == "jonkoping" {
		return fmt.Sprintf(r.baseRSSURL, "jonkopings-lan", "jonkoping")
	}
	return fmt.Sprintf(r.baseRSSURL, regionID, regionID)
}

type Region struct {
	ID   string
	Name string
	// Todo: add geometries
	// Geometry geom.T
}
