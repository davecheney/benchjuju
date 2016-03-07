package jujusvg // import "gopkg.in/juju/jujusvg.v1"

import (
	"image"
	"math"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/errgo.v1"
	"gopkg.in/juju/charm.v6-unstable"
)

// NewFromBundle returns a new Canvas that can be used
// to generate a graphical representation of the given bundle
// data. The iconURL function is used to generate a URL
// that refers to an SVG for the supplied charm URL.
// If fetcher is non-nil, it will be used to fetch icon
// contents for any icons embedded within the charm,
// allowing the generated bundle to be self-contained. If fetcher
// is nil, a default fetcher which refers to icons by their
// URLs as svg <image> tags will be used.
func NewFromBundle(b *charm.BundleData, iconURL func(*charm.URL) string, fetcher IconFetcher) (*Canvas, error) {
	if fetcher == nil {
		fetcher = &LinkFetcher{
			IconURL: iconURL,
		}
	}
	iconMap, err := fetcher.FetchIcons(b)
	if err != nil {
		return nil, err
	}

	var canvas Canvas

	// Verify the bundle to make sure that all the invariants
	// that we depend on below actually hold true.
	if err := b.Verify(nil, nil); err != nil {
		return nil, errgo.Notef(err, "cannot verify bundle")
	}
	// Go through all services in alphabetical order so that
	// we get consistent results.
	serviceNames := make([]string, 0, len(b.Services))
	for name := range b.Services {
		serviceNames = append(serviceNames, name)
	}
	sort.Strings(serviceNames)
	services := make(map[string]*service)
	servicesNeedingPlacement := make(map[string]bool)
	for _, name := range serviceNames {
		serviceData := b.Services[name]
		x, xerr := strconv.ParseFloat(serviceData.Annotations["gui-x"], 64)
		y, yerr := strconv.ParseFloat(serviceData.Annotations["gui-y"], 64)
		if xerr != nil || yerr != nil {
			if serviceData.Annotations["gui-x"] == "" && serviceData.Annotations["gui-y"] == "" {
				servicesNeedingPlacement[name] = true
				x = 0
				y = 0
			} else {
				return nil, errgo.Newf("service %q does not have a valid position", name)
			}
		}
		charmID, err := charm.ParseURL(serviceData.Charm)
		if err != nil {
			// cannot actually happen, as we've verified it.
			return nil, errgo.Notef(err, "cannot parse charm %q", serviceData.Charm)
		}
		icon := iconMap[charmID.Path()]
		svc := &service{
			name:      name,
			charmPath: charmID.Path(),
			point:     image.Point{int(x), int(y)},
			iconUrl:   iconURL(charmID),
			iconSrc:   icon,
		}
		services[name] = svc
	}
	padding := image.Point{int(math.Floor(serviceBlockSize * 1.5)), int(math.Floor(serviceBlockSize * 0.5))}
	for name := range servicesNeedingPlacement {
		vertices := []image.Point{}
		for n, svc := range services {
			if !servicesNeedingPlacement[n] {
				vertices = append(vertices, svc.point)
			}
		}
		services[name].point = getPointOutside(vertices, padding)
		servicesNeedingPlacement[name] = false
	}
	for _, name := range serviceNames {
		canvas.addService(services[name])
	}
	for _, relation := range b.Relations {
		canvas.addRelation(&serviceRelation{
			serviceA: services[strings.Split(relation[0], ":")[0]],
			serviceB: services[strings.Split(relation[1], ":")[0]],
		})
	}
	return &canvas, nil
}
