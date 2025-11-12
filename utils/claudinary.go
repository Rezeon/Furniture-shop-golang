package utils

import (
	"context"
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var ctx = context.Background()

func UploadImage(filePath string, folder string) (string, string, error) {
	cld, err := cloudinary.NewFromURL(os.Getenv("CLAUDINARY_KEY"))
	if err != nil {
		return "", "", err
	}

	resp, err := cld.Upload.Upload(ctx, filePath, uploader.UploadParams{
		UniqueFilename: api.Bool(true),
		Overwrite:      api.Bool(true),
		Folder:         folder,
	})
	if err != nil {
		log.Println("Cloudinary Init Error:", err)
		return "", "", err
	}
	return resp.SecureURL, resp.PublicID, nil
}

func DeleteImage(publicID string) error {
	cld, err := cloudinary.NewFromURL(os.Getenv("CLAUDINARY_KEY"))

	if err != nil {
		return err
	}

	_, err = cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})

	if err != nil {
		return err
	}

	return nil
}
