package dtos

type CreateClassroomRequest struct {
	Name string `json:"name" binding:"required"`
}

type ClassroomResponse struct {
	ID                  uint           `json:"id"`
	Name                string         `json:"name"`
	TeacherID           uint           `json:"teacherId"`
	Teacher             *UserResponse  `json:"teacher,omitempty"`
	Teachers            []UserResponse `json:"teachers,omitempty"`
	Students            []UserResponse `json:"students,omitempty"`
	StudentCount        int            `json:"studentCount"`
	ActiveExamID        *uint          `json:"activeExamId"`
	ActiveExamCompleted bool           `json:"activeExamCompleted"`
}

type UpdateClassroomExamRequest struct {
	ActiveExamID *uint `json:"activeExamId"`
}

type AddStudentRequest struct {
	Email     string `json:"email"`
	Matricula string `json:"matricula"`
}

type AddTeacherRequest struct {
	Email string `json:"email" binding:"required"`
}
