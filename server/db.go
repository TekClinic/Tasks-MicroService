package main

import (
	"context"
	"fmt"
	"time"

    // TODO: ppb is probably short for ppb. Rename to tasks_pb, tpb, or just pb.
	ppb "github.com/TekClinic/Tasks-MicroService/tasks_protobuf"
	"github.com/uptrace/bun"
)

const yyyy_mm_dd = "2006-01-02"

// Task defines a schema of tasks.
// TODO: Check the tags, we don't actually understand what they do.
type Task struct {
	Id                int32               `bun:",pk,autoincrement" `
	Complete          bool                ``
	Title             string              `validate:"required,min=1,max=100"`
	Description       string              ``
	Expertise         string              ``
    PatientId         int32               ``
	SpecialNote       string              `validate:"max=500"`
    // These are automatically populated by bun
	CreatedAt         time.Time           `bun:",nullzero,notnull,default:current_timestamp"`
	DeletedAt         time.Time           `bun:",soft_delete,nullzero"`
}

// toGRPC returns a GRPC version of Task.
func (task Task) toGRPC() *ppb.Task {
	return &ppb.Task{
		Id:                task.Id,
		Complete:          task.Complete,
        Title:             task.Title,
        Description:       task.Description,
        Expertise:         task.Expertise,
        PatientId:         task.PatientId,
        CreatedAt:         task.CreatedAt.Format(yyyy_mm_dd),
	}
}

// taskFromGRPC returns a Task from a GRPC version.
func taskFromGRPC(task *ppb.Task) (Task, error) {
	created_at, err := time.Parse(yyyy_mm_dd, task.GetCreatedAt())
	if err != nil {
		return Task{}, fmt.Errorf("failed to parse task creation date: %w", err)
	}
	return Task{
		Id:                task.GetId(),
		Complete:          task.GetComplete(),
        Title:             task.GetTitle(),
        Description:       task.GetDescription(),
        Expertise:         task.GetExpertise(),
        PatientId:         task.GetPatientId(),
        CreatedAt:         created_at,
	}, nil
}

// createSchemaIfNotExists creates all required schemas for task microservice.
func createSchemaIfNotExists(ctx context.Context, db *bun.DB) error {
	models := []interface{}{
		(*Task)(nil),
	}

	for _, model := range models {
		if _, err := db.NewCreateTable().IfNotExists().Model(model).Exec(ctx); err != nil {
			return err
		}
	}

    /* Copied code from patients microservice. Do we need to add deleted_at?
	// Migration code. Add created_at and deleted_at columns to the task table for soft delete.
	if _, err := db.NewRaw(
		"ALTER TABLE tasks " +
			"ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now(), " +
			"ADD COLUMN IF NOT EXISTS deleted_at timestamptz, " +
			"ADD COLUMN IF NOT EXISTS needs_translator BOOLEAN, " + // <-- Add the comma
			"ALTER COLUMN needs_translator SET DEFAULT false;").Exec(ctx); err != nil {
		return err
	}
    */

    /* Search code. Also copied from patients microservice.
	// Postgres specific code. Add a text_searchable column for full-text search.
	if _, err := db.NewRaw(
		"ALTER TABLE tasks " +
			"ADD COLUMN IF NOT EXISTS text_searchable tsvector " +
			"GENERATED ALWAYS AS " +
			"(" +
			"setweight(to_tsvector('simple', coalesce(personal_id_id, '')), 'A') || " +
			"setweight(to_tsvector('simple', coalesce(phone_number, '')), 'A')   || " +
			"setweight(to_tsvector('simple', coalesce(name, '')), 'B')           || " +
			"setweight(to_tsvector('simple', coalesce(special_note, '')), 'C')   || " +
			"setweight(to_tsvector('simple', coalesce(referred_by, '')), 'D')" +
			") STORED").Exec(ctx); err != nil {
		return err
	}
    */

	return nil
}
