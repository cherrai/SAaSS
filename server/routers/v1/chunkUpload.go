package routerV1

import (
	controllersV1 "github.com/cherrai/SAaSS/controllers/v1"
	"github.com/cherrai/SAaSS/services/middleware"
)

func (r Routerv1) InitChunkUpload() {
	chunkUpload := new(controllersV1.ChunkUploadController)

	role := middleware.RoleMiddlewareOptions{
		BaseUrl: r.BaseUrl,
	}

	r.Group.POST(
		role.SetRole("/chunkupload/create", &middleware.RoleOptionsType{
			CheckApp:           true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		chunkUpload.CreateChunk)
	r.Group.POST(
		role.SetRole("/chunkupload/upload", &middleware.RoleOptionsType{
			CheckApp:           false,
			Authorize:          true,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		chunkUpload.UploadChunk)

}
