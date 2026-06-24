import { api, setAuthToken } from "./client";
import type { components } from "./schema";

export type AuthResponse = components["schemas"]["AuthResponse"];
export type CurrentUserResponse = components["schemas"]["CurrentUserResponse"];
export type Department = components["schemas"]["Department"];
export type Employee = components["schemas"]["Employee"];
export type Location = components["schemas"]["Location"];
export type Role = components["schemas"]["Role"];
export type Shift = components["schemas"]["Shift"];
export type LiveViewEntry = components["schemas"]["LiveViewEntry"];
export type LeaveRequest = components["schemas"]["LeaveRequest"];
export type LeaveStatus = components["schemas"]["LeaveStatus"];
export type Tour = components["schemas"]["Tour"];
export type TourCategory = components["schemas"]["TourCategory"];
export type CreateTourRequest = components["schemas"]["CreateTourRequest"];
export type UpdateTourRequest = components["schemas"]["UpdateTourRequest"];

export type CreateDepartmentRequest =
  components["schemas"]["CreateDepartmentRequest"];
export type CreateEmployeeRequest = components["schemas"]["CreateEmployeeRequest"];
export type CreateLocationRequest = components["schemas"]["CreateLocationRequest"];
export type CreateRoleRequest = components["schemas"]["CreateRoleRequest"];
export type CreateShiftRequest = components["schemas"]["CreateShiftRequest"];
export type UpdateEmployeeRequest = components["schemas"]["UpdateEmployeeRequest"];
export type UpdateShiftRequest = components["schemas"]["UpdateShiftRequest"];
export type UpdateDepartmentRequest = components["schemas"]["UpdateDepartmentRequest"];
export type UpdateRoleRequest = components["schemas"]["UpdateRoleRequest"];
export type LoginRequest = components["schemas"]["LoginRequest"];
export type SignupRequest = components["schemas"]["SignupRequest"];

export { setAuthToken };

type ApiResult<T> = {
  data?: T;
  error?: unknown;
  response: Response;
};

function errorMessage(error: unknown, fallback: string): string {
  if (typeof error === "object" && error && "message" in error) {
    const message = (error as { message?: unknown }).message;
    if (typeof message === "string" && message.trim()) {
      return message;
    }
  }
  return fallback;
}

async function unwrap<T>(result: ApiResult<T>, fallback: string): Promise<T> {
  if (result.error || !result.response.ok) {
    throw new Error(errorMessage(result.error, fallback));
  }
  if (result.data === undefined) {
    throw new Error(fallback);
  }
  return result.data;
}

async function unwrapEmpty(
  result: ApiResult<unknown>,
  fallback: string,
): Promise<void> {
  if (result.error || !result.response.ok) {
    throw new Error(errorMessage(result.error, fallback));
  }
}

export const authApi = {
  async login(body: LoginRequest): Promise<AuthResponse> {
    return unwrap(
      await api.POST("/auth/login", { body }),
      "Unable to log in.",
    );
  },
  async signup(body: SignupRequest): Promise<AuthResponse> {
    return unwrap(
      await api.POST("/auth/signup", { body }),
      "Unable to create tenant.",
    );
  },
  async me(): Promise<CurrentUserResponse> {
    return unwrap(await api.GET("/auth/me"), "Unable to load current user.");
  },
};

export const departmentsApi = {
  async list(): Promise<Department[]> {
    return unwrap(
      await api.GET("/departments"),
      "Unable to load departments.",
    );
  },
  async create(body: CreateDepartmentRequest): Promise<Department> {
    return unwrap(
      await api.POST("/departments", { body }),
      "Unable to create department.",
    );
  },
  async update(id: string, body: UpdateDepartmentRequest): Promise<Department> {
    return unwrap(
      await api.PATCH("/departments/{id}", { params: { path: { id } }, body }),
      "Unable to update department.",
    );
  },
  async delete(id: string): Promise<void> {
    await unwrapEmpty(
      await api.DELETE("/departments/{id}", { params: { path: { id } } }),
      "Unable to delete department.",
    );
  },
};

export const employeesApi = {
  async list(filters: {
    department_id?: string;
    status?: Employee["status"];
  } = {}): Promise<Employee[]> {
    return unwrap(
      await api.GET("/employees", { params: { query: filters } }),
      "Unable to load employees.",
    );
  },
  async create(body: CreateEmployeeRequest): Promise<Employee> {
    return unwrap(
      await api.POST("/employees", { body }),
      "Unable to create employee.",
    );
  },
  async update(id: string, body: UpdateEmployeeRequest): Promise<Employee> {
    return unwrap(
      await api.PATCH("/employees/{id}", { params: { path: { id } }, body }),
      "Unable to update employee.",
    );
  },
  async delete(id: string): Promise<void> {
    await unwrapEmpty(
      await api.DELETE("/employees/{id}", { params: { path: { id } } }),
      "Unable to delete employee.",
    );
  },
};

export const rolesApi = {
  async list(): Promise<Role[]> {
    return unwrap(await api.GET("/roles"), "Unable to load roles.");
  },
  async create(body: CreateRoleRequest): Promise<Role> {
    return unwrap(
      await api.POST("/roles", { body }),
      "Unable to create role.",
    );
  },
  async update(id: string, body: UpdateRoleRequest): Promise<Role> {
    return unwrap(
      await api.PATCH("/roles/{id}", { params: { path: { id } }, body }),
      "Unable to update role.",
    );
  },
  async delete(id: string): Promise<void> {
    await unwrapEmpty(
      await api.DELETE("/roles/{id}", { params: { path: { id } } }),
      "Unable to delete role.",
    );
  },
};

export const locationsApi = {
  async list(): Promise<Location[]> {
    return unwrap(await api.GET("/locations"), "Unable to load locations.");
  },
  async create(body: CreateLocationRequest): Promise<Location> {
    return unwrap(
      await api.POST("/locations", { body }),
      "Unable to create location.",
    );
  },
  async delete(id: string): Promise<void> {
    await unwrapEmpty(
      await api.DELETE("/locations/{id}", { params: { path: { id } } }),
      "Unable to delete location.",
    );
  },
};

export const shiftsApi = {
  async list(filters: {
    employee_id?: string;
    status?: Shift["status"];
    from?: string;
    to?: string;
  } = {}): Promise<Shift[]> {
    return unwrap(
      await api.GET("/shifts", { params: { query: filters } }),
      "Unable to load shifts.",
    );
  },
  async create(body: CreateShiftRequest): Promise<Shift> {
    return unwrap(
      await api.POST("/shifts", { body }),
      "Unable to create shift.",
    );
  },
  async update(id: string, body: UpdateShiftRequest): Promise<Shift> {
    return unwrap(
      await api.PATCH("/shifts/{id}", { params: { path: { id } }, body }),
      "Unable to update shift.",
    );
  },
  async publish(id: string): Promise<Shift> {
    return unwrap(
      await api.POST("/shifts/{id}/publish", { params: { path: { id } } }),
      "Unable to publish shift.",
    );
  },
  async delete(id: string): Promise<void> {
    await unwrapEmpty(
      await api.DELETE("/shifts/{id}", { params: { path: { id } } }),
      "Unable to delete shift.",
    );
  },
};

// Manager live view ("who's working now"): published shifts in a window joined
// with current attendance state. Window defaults to the current UTC day when
// from/to are omitted; pass local-day bounds for correct timezone behavior.
export const liveApi = {
  async list(filters: { from?: string; to?: string } = {}): Promise<LiveViewEntry[]> {
    return unwrap(
      await api.GET("/live", { params: { query: filters } }),
      "Unable to load the live view.",
    );
  },
};

// Manager time-off review: list all leave requests and approve/reject them.
export const leaveApi = {
  async list(
    filters: { status?: LeaveStatus; employee_id?: string } = {},
  ): Promise<LeaveRequest[]> {
    return unwrap(
      await api.GET("/leave", { params: { query: filters } }),
      "Unable to load leave requests.",
    );
  },
  async approve(id: string): Promise<LeaveRequest> {
    return unwrap(
      await api.POST("/leave/{id}/approve", { params: { path: { id } } }),
      "Unable to approve the request.",
    );
  },
  async reject(id: string): Promise<LeaveRequest> {
    return unwrap(
      await api.POST("/leave/{id}/reject", { params: { path: { id } } }),
      "Unable to reject the request.",
    );
  },
};

export const toursApi = {
  async list(
    filters: { category?: TourCategory; active?: boolean } = {},
  ): Promise<Tour[]> {
    return unwrap(
      await api.GET("/tours", { params: { query: filters } }),
      "Unable to load tours.",
    );
  },
  async create(body: CreateTourRequest): Promise<Tour> {
    return unwrap(await api.POST("/tours", { body }), "Unable to create tour.");
  },
  async update(id: string, body: UpdateTourRequest): Promise<Tour> {
    return unwrap(
      await api.PATCH("/tours/{id}", { params: { path: { id } }, body }),
      "Unable to update tour.",
    );
  },
  async delete(id: string): Promise<void> {
    await unwrapEmpty(
      await api.DELETE("/tours/{id}", { params: { path: { id } } }),
      "Unable to delete tour.",
    );
  },
};
