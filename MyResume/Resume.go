package main

import (
	"fmt"
)

type Resume struct {
	Name         string
	Phone        string
	Email        string
	Location     string
	Summary      string
	Skills       []string
	Projects     []Project
	Interships   []Intership
	Educations   []Education
	Certifiactes []string
	Links        []Link
}

type Project struct {
	Title       string
	Description string
}
type Intership struct {
	CompanyName string
	Location    string
	Role        string
	StartDate   string
	EndDate     string
}
type Education struct {
	CollegeName string
	Location    string
	Degree      string
	Graduation  string
	CGPA        float64
}
type Link struct {
	WebsiteName string
	LinkURL     string
}

func printResume(r Resume) {
	fmt.Println("Name:", r.Name)
	fmt.Println("Phone:", r.Phone)
	fmt.Println("Email:", r.Email)
	fmt.Println("Location:", r.Location)
	fmt.Println("\nSummary:\n", r.Summary)

	fmt.Println("\nSkills:")
	for _, skill := range r.Skills {
		fmt.Println(" -", skill)
	}

	fmt.Println("\nProjects:")
	for _, project := range r.Projects {
		fmt.Println(" Title:", project.Title)
		fmt.Println(" Description:", project.Description)
		fmt.Println()
	}

	fmt.Println("Internships:")
	for _, intern := range r.Interships {
		fmt.Println(" Company:", intern.CompanyName)
		fmt.Println(" Location:", intern.Location)
		fmt.Println(" Role:", intern.Role)
		fmt.Println(" Duration:", intern.StartDate, "to", intern.EndDate)
		fmt.Println()
	}

	fmt.Println("Education:")
	for _, edu := range r.Educations {
		fmt.Println(" College:", edu.CollegeName)
		fmt.Println(" Location:", edu.Location)
		fmt.Println(" Degree:", edu.Degree)
		fmt.Println(" Graduation:", edu.Graduation)
		fmt.Printf(" CGPA: %.2f\n\n", edu.CGPA)
	}

	fmt.Println("Certificates:")
	for _, cert := range r.Certifiactes {
		fmt.Println(" -", cert)
	}

	fmt.Println("\nLinks:")
	for _, link := range r.Links {
		fmt.Println(" ", link.WebsiteName, link.LinkURL)
	}
}

func main() {
	myResume := Resume{
		Name:     "Vishnu Balaji T",
		Phone:    "09361428953",
		Email:    "tjvishnu.vj@gmail.com",
		Location: "Madurai, Tamil Nadu",
		Summary: `An enthusiastic fresher with highly motivated and leadership skills having bachelor's degree in Information Technology.
Seeking a responsible career opportunity to fully utilize my training and skills,while making a significant contribution to the
success of the company.`,
		Skills: []string{"MySQL", "Java", "HTML5", "CSS", "Spring Boot", "Git"},
		Projects: []Project{
			{
				Title:       "Billing System Application",
				Description: "Developed a role-based billing system in Go that handles sales, inventory, GST, and expiry checks using MySQL.",
			},
			{
				Title:       "Login/Signup with JWT Authentication",
				Description: "Built secure authentication using JWT and bcrypt with Gorilla Mux and MySQL backend.",
			},
			{
				Title:       "Email Scheduler",
				Description: "Automated email scheduler using gomail and cron to send emails at specific intervals.",
			},
			{
				Title:       "URL Shortener",
				Description: "Built a REST API to shorten URLs and handle redirection using a unique code.",
			},
		},
		Interships: []Intership{
			{
				CompanyName: "Phoenix Softtech",
				Location:    "Madurai,TamilNadu.",
				Role:        "Software Intern",
				StartDate:   "01/2025",
				EndDate:     "02/2025",
			},
		},
		Educations: []Education{
			{
				CollegeName: "P.S.N.A College Of Engineering And Technology Dindigul",
				Location:    "Dindigul,TamilNadu.",
				Degree:      "B.Tech Information Technology",
				Graduation:  "04/2023",
				CGPA:        8.34,
			},
		},
		Certifiactes: []string{
			"Fundamentals of Java Programming(Coursera)",
		},
		Links: []Link{
			{
				WebsiteName: "GitHub:",
				LinkURL:     "https://github.com/VishnuBalaji07",
			},
			{
				WebsiteName: "LinkedIn:",
				LinkURL:     "https://www.linkedin.com/in/vishnu-balaji-93183428a/",
			},
		},
	}

	printResume(myResume)
}
