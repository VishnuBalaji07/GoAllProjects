package main

import (
	"fmt"
)

type Employee struct {
	ID         int
	Name       string
	Position   string
	Salary     int
	Phone_no   int
	Work_Shift string
	Role       string
}

func getEmployeeById(employees map[int]Employee, id int) (*Employee, bool) {
	emp, found := employees[id]
	if found {
		return &emp, true
	}
	return nil, false
}
func main() {
	employees := map[int]Employee{
		101: {ID: 101, Name: "Alice", Position: "Developer", Salary: 75000, Phone_no: 1234567890, Work_Shift: "Day", Role: "Backend"},
		102: {ID: 102, Name: "Bob", Position: "Designer", Salary: 65000, Phone_no: 2345678901, Work_Shift: "Evening", Role: "UI/UX"},
		103: {ID: 103, Name: "Charlie", Position: "Manager", Salary: 85000, Phone_no: 3456789012, Work_Shift: "Day", Role: "Team Lead"},
		104: {ID: 104, Name: "Diana", Position: "QA Engineer", Salary: 60000, Phone_no: 4567890123, Work_Shift: "Night", Role: "Automation"},
	}
	var idToFind int
	fmt.Print("Enter Employee ID to search:")
	_, err := fmt.Scanln(&idToFind)
	if err != nil {
		fmt.Println("Invalid input. Please enter a valid number.")
		return
	}
	if emp, found := getEmployeeById(employees, idToFind); found {
		fmt.Printf("\nEmployee Found:\n")
		fmt.Printf("ID: %d\nName: %s\nPosition: %s\nSalary: %s\nPhone: %d\nShift: %s\nRole: %s\n",
			emp.ID, emp.Name, emp.Position, emp.Salary, emp.Phone_no, emp.Work_Shift, emp.Role)
	} else {
		fmt.Printf("Employee with ID %d not found.\n", idToFind)
	}

}
