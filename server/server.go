package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"

	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"

	ms "github.com/TekClinic/MicroService-Lib"
	ppb "github.com/TekClinic/Tasks-MicroService/tasks_protobuf"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// tasksServer is an implementation of GRPC task microservice. It provides access to a database via db field.
type tasksServer struct {
	ppb.UnimplementedTasksServiceServer
	ms.BaseServiceServer
	db *bun.DB
	// use a single instance of Validate, it caches struct info
	validate *validator.Validate
}

const (
	envDBAddress  = "DB_ADDR"
	envDBUser     = "DB_USER"
	envDBDatabase = "DB_DATABASE"
	envDBPassword = "DB_PASSWORD"

	applicationName = "tasks"

	permissionDeniedMessage = "You don't have enough permission to access this resource"

	maxPaginationLimit = 50
)

// GetTask returns a task that corresponds to the given id.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If a task with a given id doesn't exist, codes.NotFound is returned.
func (server tasksServer) GetTask(ctx context.Context, req *ppb.GetTaskRequest) (
	*ppb.GetTaskResponse, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	task := new(Task)
	err = server.db.NewSelect().
		Model(task).
		Where("? = ?", bun.Ident("id"), req.GetId()).
		WhereAllWithDeleted().
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "task is not found")
		}
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch a tasks by id: %w", err).Error())
	}
	return &ppb.GetTaskResponse{Task: task.toGRPC()}, nil
}

// GetTasksIDs returns a list of tasks' ids with given filters and pagination.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// Offset value is used for pagination. Required be a non-negative value.
// Limit value is used for pagination. Required to be a positive value.
func (server tasksServer) GetTasksIDs(ctx context.Context,
	req *ppb.GetTasksIDsRequest) (*ppb.GetTasksIDsResponse, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	if req.GetOffset() < 0 {
		return nil, status.Error(codes.InvalidArgument, "offset has to be a non-negative integer")
	}
	if req.GetLimit() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "limit has to be a positive integer")
	}
	if req.GetLimit() > maxPaginationLimit {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("maximum allowed limit values is %d", maxPaginationLimit))
	}

	var ids []int32
	baseQuery := server.db.NewSelect().Model((*Task)(nil)).Column("id")

    // TODO: Implement search
    /*
	if req.GetSearch() != "" {
		// Postgres specific code. Use full-text search to search for tasks.
		baseQuery = baseQuery.
			TableExpr("replace(websearch_to_tsquery('simple', ?)::text || ' ',''' ',''':*') query", req.GetSearch()).
			Where("text_searchable @@ query::tsquery", req.GetSearch()).
			OrderExpr("ts_rank(text_searchable, query::tsquery) DESC")
	}
    */

	err = baseQuery.
		Offset(int(req.GetOffset())).
		Limit(int(req.GetLimit())).
		Scan(ctx, &ids)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch tasks: %w", err).Error())
	}
	count, err := baseQuery.Count(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to count tasks: %w", err).Error())
	}

	return &ppb.GetTasksIDsResponse{
		Count:   int32(count),
		Results: ids,
	}, nil
}

// CreateTask creates a task with the given specifications.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If some argument is missing or not valid, codes.InvalidArgument is returned.
func (server tasksServer) CreateTask(ctx context.Context,
	req *ppb.CreateTaskRequest) (*ppb.CreateTaskResponse, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	task := Task{
		Complete:       false,
        Title:          req.GetTitle(),
        Description:    req.GetDescription(),
        Expertise:      req.GetExpertise(),
        PatientId:      req.GetPatientId(),
	}
	if err = server.validate.Struct(task); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err = server.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		// insert the task itself
		if _, txErr := tx.NewInsert().Model(&task).Exec(ctx); txErr != nil {
			return txErr
		}
		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to create a task: %w", err).Error())
	}
	return &ppb.CreateTaskResponse{Id: task.Id}, nil
}

// DeleteTask deletes a task with the given id.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If a task with a given id doesn't exist, codes.NotFound is returned.
func (server tasksServer) DeleteTask(ctx context.Context, req *ppb.DeleteTaskRequest) (
	*ppb.DeleteTaskResponse, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	res, err := server.db.NewDelete().Model((*Task)(nil)).Where("id = ?", req.GetId()).Exec(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to delete a task: %w", err).Error())
	}
	// if db supports affected rows count and no rows were affected, return not found
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return nil, status.Error(codes.NotFound, "task is not found")
	}
	return &ppb.DeleteTaskResponse{}, nil
}

// UpdateTask updates a task with the given id and data.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If some argument is missing or not valid, codes.InvalidArgument is returned.
// If a task with a given id doesn't exist, codes.NotFound is returned.
func (server tasksServer) UpdateTask(ctx context.Context, req *ppb.UpdateTaskRequest) (
	*ppb.UpdateTaskResponse, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	task, err := taskFromGRPC(req.GetTask())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err = server.validate.Struct(task); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if task.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "Task ID is required")
	}


	if err = server.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		// update the task
		res, txErr := tx.NewUpdate().
			Model(&task).
			ExcludeColumn("created_at", "deleted_at").
			WherePK().
			Exec(ctx)
		if txErr != nil {
			return status.Error(codes.Internal, fmt.Errorf("failed to update a task: %w", txErr).Error())
		}

		// if db supports affected rows count and no rows were affected, return not found
		rows, rowsErr := res.RowsAffected()
		if rowsErr == nil && rows == 0 {
			return status.Error(codes.NotFound, "task is not found")
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return &ppb.UpdateTaskResponse{Id: task.Id}, nil
}

// GetTasksByPatient returns a list of tasks for a given patient id.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If no tasks are found for the patient, an empty list is returned.
func (server tasksServer) GetTasksByPatient(ctx context.Context, req *ppb.GetTasksByPatientRequest) (*ppb.GetTasksByPatientResponse, error) {
    claims, err := server.VerifyToken(ctx, req.GetToken())
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, err.Error())
    }
    if !claims.HasRole("admin") {
        return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
    }

    var tasks []Task
    err = server.db.NewSelect().
        Model(&tasks).
        Where("patient_id = ?", req.GetPatientId()).
        Scan(ctx)
    if err != nil {
        return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch tasks: %w", err).Error())
    }

    grpcTasks := make([]*ppb.Task, len(tasks))
    for i, t := range tasks {
        grpcTasks[i] = t.toGRPC()
    }

    return &ppb.GetTasksByPatientResponse{
        Tasks: grpcTasks,
    }, nil
}

// createTasksServer initializes a tasksServer with all the necessary fields.
func createTasksServer() (*tasksServer, error) {
	base, err := ms.CreateBaseServiceServer()
	if err != nil {
		return nil, err
	}
	addr, err := ms.GetRequiredEnv(envDBAddress)
	if err != nil {
		return nil, err
	}
	user, err := ms.GetRequiredEnv(envDBUser)
	if err != nil {
		return nil, err
	}
	password, err := ms.GetRequiredEnv(envDBPassword)
	if err != nil {
		return nil, err
	}
	database, err := ms.GetRequiredEnv(envDBDatabase)
	if err != nil {
		return nil, err
	}
	connector := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(addr),
		pgdriver.WithUser(user),
		pgdriver.WithPassword(password),
		pgdriver.WithDatabase(database),
		pgdriver.WithApplicationName(applicationName),
		pgdriver.WithInsecure(!ms.HasSecureConnection()),
	)
	db := bun.NewDB(sql.OpenDB(connector), pgdialect.New())
	db.AddQueryHook(ms.GetDBQueryHook())
	return &tasksServer{
		BaseServiceServer: base,
		db:                db,
		validate:          validator.New(validator.WithRequiredStructEnabled())}, nil
}

func main() {
	service, err := createTasksServer()
	if err != nil {
		zap.L().Fatal("Failed to create a task server", zap.Error(err))
	}

	err = createSchemaIfNotExists(context.Background(), service.db)
	if err != nil {
		zap.L().Fatal("Failed to create a schema", zap.Error(err))
	}

	listen, err := net.Listen("tcp", ":"+service.GetPort())
	if err != nil {
		zap.L().Fatal("Failed to listen", zap.Error(err))
	}

	srv := grpc.NewServer(ms.GetGRPCServerOptions()...)
	ppb.RegisterTasksServiceServer(srv, service)

	zap.L().Info("Server listening on :" + service.GetPort())
	if err = srv.Serve(listen); err != nil {
		zap.L().Fatal("Failed to serve", zap.Error(err))
	}
}
