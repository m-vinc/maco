package engine

import (
	"context"
	"sort"

	"github.com/m-vinc/maco/pkg/image"
)

type CatalogImage struct {
	ID          string `json:"id"`
	Distro      string `json:"distro"`
	Version     string `json:"version"`
	Arch        string `json:"arch"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Downloaded  bool   `json:"downloaded"`
	SizeBytes   int64  `json:"size_bytes"`
}

func (e *Engine) Catalog() []CatalogImage {
	dir := e.paths.ImagesDir()
	result := make([]CatalogImage, 0, len(image.Catalog))
	for _, img := range image.Catalog {
		size, downloaded := image.CachedSize(dir, img)
		result = append(result, CatalogImage{
			ID:          img.Name,
			Distro:      img.Distro,
			Version:     img.Version,
			Arch:        img.Arch,
			DisplayName: img.DisplayName,
			Description: img.Description,
			Downloaded:  downloaded,
			SizeBytes:   size,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].DisplayName < result[j].DisplayName })
	return result
}

func (e *Engine) DownloadCatalogImage(ctx context.Context, id string) error {
	img, err := image.Lookup(id)
	if err != nil {
		return err
	}
	_, err = image.PullContext(ctx, e.paths.ImagesDir(), img)
	return err
}

func (e *Engine) DeleteCatalogImage(id string) error {
	img, err := image.Lookup(id)
	if err != nil {
		return err
	}
	return image.RemoveCached(e.paths.ImagesDir(), img)
}
