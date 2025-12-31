package mappers

import (
	"time"

	dto "github.com/create-go-app/fiber-go-template/app/dto"
	entities "github.com/create-go-app/fiber-go-template/app/entities"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.1 gen ./...

// goverter:converter
// goverter:output:package generated
// goverter:extend CopyTime
type TaskMapper interface {
	// goverter:map Creator.Name CreateByName
	// goverter:map Assignee.Name AssignedToName
	EntityToRes(source entities.Task) dto.TaskRes

	EntitiesToResList(source []entities.Task) []dto.TaskRes

	// Map request to entity (service will fill missing fields)
	// goverter:ignore ID Status CreatedBy CreatedAt UpdatedAt Creator Assignee
	CreateReqToEntity(source dto.CreateTaskReq) entities.Task
}

func CopyTime(t time.Time) time.Time { return t }
