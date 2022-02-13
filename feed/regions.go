package feed

type region struct {
	ID   string
	Name string
	// Todo: add geometries
	// Geometry geom.T
}

var regions = map[string]region{
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

// func (r *Regions) GetRSSURL(regionID string) string {
// 	if regionID == "jonkoping" {
// 		return fmt.Sprintf(r.baseRSSURL, "jonkopings-lan", "jonkoping")
// 	}
// 	return fmt.Sprintf(r.baseRSSURL, regionID, regionID)
// }
