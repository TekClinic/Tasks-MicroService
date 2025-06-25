## gRPC Functions

### GetTask

Retrieves the details of a specific task by its ID.

**Request:**

```protobuf
message GetTaskRequest {
  string token = 1; // Authentication token
  int32 id = 2;     // ID of the task
}
```

**Response:**

```protobuf
message GetTaskResponse {
  Task task = 1; // Details of the task
}
```

**Errors:**

- `Unauthenticated` - Token is not valid or expired.
- `PermissionDenied` - Token is not authorized with the *admin* role.
- `NotFound` - Task with the given ID does not exist.

---

### GetTasksIDs

Retrieves a paginated list of task IDs with optional search filtering.

**Request:**

```protobuf
message GetTasksIDsRequest {
  string token = 1;  // Authentication token
  int32 limit = 2;   // Maximum number of results
  int32 offset = 3;  // Pagination offset
  string search = 4; // Optional search term (e.g., title, description)
}
```

**Response:**

```protobuf
message GetTasksIDsResponse {
  int32 count = 1;           // Total number of matching tasks
  repeated int32 results = 2; // List of task IDs
}
```

**Errors:**

- `Unauthenticated` - Token is not valid or expired.
- `PermissionDenied` - Token is not authorized with the *admin* role.
- `InvalidArgument` - `offset`, `limit`, or `search` parameters are invalid.

---

### CreateTask

Creates a new task and assigns it to a patient.

**Request:**

```protobuf
message CreateTaskRequest {
  string token = 1;         // Authentication token
  string title = 2;         // Title of the task
  string description = 3;   // Description of the task
  string expertise = 4;     // Area of expertise related to the task
  int32 patient_id = 5;     // ID of the patient this task belongs to
}
```

**Response:**

```protobuf
message CreateTaskResponse {
  int32 id = 1; // ID of the newly created task
}
```

**Errors:**

- `Unauthenticated` - Token is not valid or expired.
- `PermissionDenied` - Token is not authorized with the *admin* role.
- `InvalidArgument` - Missing or malformed task information.
- `NotFound` - Patient with the given ID does not exist.

---

### DeleteTask

Deletes a task by its ID.

**Request:**

```protobuf
message DeleteTaskRequest {
  string token = 1; // Authentication token
  int32 id = 2;     // ID of the task to delete
}
```

**Response:**

```protobuf
message DeleteTaskResponse {}
```

**Errors:**

- `Unauthenticated` - Token is not valid or expired.
- `PermissionDenied` - Token is not authorized with the *admin* role.
- `NotFound` - Task with the given ID does not exist.

---

### UpdateTask

Updates the details of an existing task.

**Request:**

```protobuf
message UpdateTaskRequest {
  string token = 1; // Authentication token
  Task task = 2;    // Updated task details
}
```

**Response:**

```protobuf
message UpdateTaskResponse {
  int32 id = 1; // ID of the updated task
}
```

**Errors:**

- `Unauthenticated` - Token is not valid or expired.
- `PermissionDenied` - Token is not authorized with the *admin* role.
- `InvalidArgument` - Malformed or incomplete task data.
- `NotFound` - Task with the given ID does not exist.

---
### GetTasksByPatient

Retrieves all tasks assigned to a specific patient.

**Request:**

```protobuf
message GetTasksByPatientRequest {
  string token = 1;     // Authentication token
  int32 patient_id = 2; // ID of the patient
}
```

**Response:**

```protobuf
message GetTasksByPatientResponse {
  repeated Task tasks = 1; // List of tasks for the patient
}
```

**Errors:**
- `Unauthenticated` – Token is not valid or expired.
- `PermissionDenied` – Token is not authorized.
- `NotFound` – Patient with the given ID does not exist or has no tasks.

## Model Definition

```protobuf
message Task {
  int32 id = 1;             // Task ID
  bool complete = 2;        // Task completion status
  string title = 3;         // Title of the task
  string description = 4;   // Description of the task
  string expertise = 5;     // Area of expertise required
  int32 patient_id = 6;     // ID of the associated patient
  string created_at = 7;    // Timestamp of task creation (RFC3339 format)
}
```
