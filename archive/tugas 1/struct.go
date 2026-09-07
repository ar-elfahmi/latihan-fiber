//go:build ignore

package main

import "fmt"

// Struct Student
type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

// Value receiver — hanya membaca, tidak mengubah data
func (s Student) GetInfo() string {
	status := "Nonaktif"
	if s.IsActive {
		status = "Aktif"
	}
	return fmt.Sprintf("ID: %d | Nama: %s | Nilai: %.2f | Status: %s", s.ID, s.Name, s.Grade, status)
}

// Pointer receiver — mengubah data asli
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

func (s *Student) Activate() {
	s.IsActive = true
}

func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	s := Student{ID: 1, Name: "Budi", Grade: 3.50, IsActive: false}

	fmt.Println("Awal     :", s.GetInfo())

	s.Activate()
	fmt.Println("Aktivasi :", s.GetInfo())

	s.UpdateGrade(3.85)
	fmt.Println("Update   :", s.GetInfo())

	s.Deactivate()
	fmt.Println("Nonaktif :", s.GetInfo())
}
