package dtos

type CreateClassroomRequest struct {
	Name string `json:"name" binding:"required"`
}

type ClassroomResponse struct {
	ID           uint           `json:"id"`
	Name         string         `json:"name"`
	TeacherID    uint           `json:"teacherId"`
	Teacher      *UserResponse  `json:"teacher,omitempty"` // Use pointer to omit if nil
	Students     []UserResponse `json:"students,omitempty"`
	StudentCount int            `json:"studentCount"`
}

type AddStudentRequest struct {
	Email string `json:"email" binding:"required,email"`
}
