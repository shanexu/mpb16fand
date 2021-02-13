package main

import (
	"fmt"
	"log"
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"github.com/pelletier/go-toml"

	"github.com/shanexu/mbp16fanctl/pkg/config"
	"github.com/shanexu/mbp16fanctl/pkg/sensors"
)

type Mbp16fand struct {
}

func (m *Mbp16fand) getTemp(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("temp-name")
	s, ok := sensors.TempSensors[name]
	if !ok {
		response.WriteErrorString(http.StatusNotFound, "Temp sensor could not be found.")
	} else {
		response.WriteEntity(s)
	}
}

func (m *Mbp16fand) findTemps(request *restful.Request, response *restful.Response) {
	response.WriteEntity(sensors.TempSensors)
}

func (m *Mbp16fand) findFans(request *restful.Request, response *restful.Response) {
	response.WriteEntity(sensors.FanSensors)
}

func (m *Mbp16fand) getFan(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("fan-name")
	s, ok := sensors.FanSensors[name]
	if !ok {
		response.WriteErrorString(http.StatusNotFound, "Fan sensor could not be found.")
	} else {
		response.WriteEntity(s)
	}
}

func (m *Mbp16fand) WebService() []*restful.WebService {
	var wss []*restful.WebService
	ws := new(restful.WebService)
	ws.
		Path("/temp").
		Consumes(restful.MIME_JSON, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_JSON)

	ws.Route(ws.GET("/").To(m.findTemps).
		Doc("find temp sensors").
		Metadata(restfulspec.KeyOpenAPITags, []string{"temp"}).
		Writes(map[string]sensors.TempSensor{}).
		Returns(http.StatusOK, "OK", map[string]sensors.TempSensor{}))

	ws.Route(ws.GET("/{temp-name}").To(m.getTemp).
		Doc("get a temp sensor").
		Param(ws.PathParameter("temp-name", "temp sensor name").DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, []string{"temp"}).
		Writes(sensors.TempSensor{}).
		Returns(http.StatusOK, "OK", sensors.TempSensor{}))
	wss = append(wss, ws)
	ws = new(restful.WebService)
	ws.Path("/fan").
		Consumes(restful.MIME_JSON, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_JSON)

	ws.Route(ws.GET("/").To(m.findFans).
		Doc("find fan sensors").
		Metadata(restfulspec.KeyOpenAPITags, []string{"fan"}).
		Writes(map[string]sensors.FanSensor{}).
		Returns(http.StatusOK, "OK", map[string]sensors.FanSensor{}))

	ws.Route(ws.GET("/{fan-name}").To(m.getFan).
		Doc("get a fan sensor").
		Param(ws.PathParameter("fan-name", "fan sensor name").DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, []string{"fan"}).
		Writes(sensors.FanSensor{}).
		Returns(http.StatusOK, "OK", sensors.FanSensor{}))
	wss = append(wss, ws)
	return wss
}

func main() {
	m := &Mbp16fand{}
	wss := m.WebService()
	for _, ws := range wss {
		restful.DefaultContainer.Add(ws)
	}

	cfg := restfulspec.Config{
		WebServices:                   restful.RegisteredWebServices(), // you control what services are visible
		APIPath:                       "/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(cfg))
	http.Handle("/apidocs/", http.StripPrefix("/apidocs/", http.FileServer(http.Dir("./swagger-ui"))))

	t, err := toml.LoadFile("/etc/mbp16fand/mbp16fand.toml")
	if err != nil {
		panic(err)
	}
	fmt.Println(t)

	c := config.Config{}
	t.Unmarshal(&c)
	profile, ok := c.Profiles[c.ActiveProfile]
	if !ok {
		panic(fmt.Sprintf("no such profile %q", c.ActiveProfile))
	}
	for _, s := range sensors.FanSensors {
		p := &sensors.Profile{
			Manual: profile.Manual,
			Speed:  profile.Speed,
		}
		if profile.Speed == 0 && profile.Manual {
			if profile.HighTemp >= 0 && profile.LowTemp >= 0 && profile.HighTemp > profile.LowTemp {
				k := float64(s.Max-s.Min) / float64(profile.HighTemp-profile.LowTemp)
				c := float64(s.Min) - k*float64(profile.LowTemp)
				p.K = k
				p.C = c
			} else {
				panic("high temp and low temp not valid")
			}
		}

		s.Profile = p
	}
	sensors.NotifyFanSensors()
	log.Printf("start listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "TempSensorService",
			Description: "Resource for managing TempSensors",
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "shanexu",
					Email: "xusheng0711@gmail.com",
					URL:   "https://xusheng.org",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "MIT",
					URL:  "http://mit.org",
				},
			},
			Version: "1.0.0",
		},
	}
	swo.Tags = []spec.Tag{
		{
			TagProps: spec.TagProps{
				Name:        "users",
				Description: "Managing users",
			},
		},
	}
}
