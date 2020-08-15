package main

import (
	"context"
	"flag"
	"strconv"

	"github.com/andrewmthomas87/nuapiclient"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
)

var (
	runCourses     string
	runTerms       bool
	runSchools     bool
	runSubjects    bool
	runInstructors bool
	runBuildings   bool
	runRooms       bool
)

func main() {
	flag.StringVar(&runCourses, "c", "", "run courses")
	flag.BoolVar(&runTerms, "t", false, "run terms")
	flag.BoolVar(&runSchools, "sc", false, "run schools")
	flag.BoolVar(&runSubjects, "su", false, "run subjects")
	flag.BoolVar(&runInstructors, "i", false, "run instructors")
	flag.BoolVar(&runBuildings, "b", false, "run buildings")
	flag.BoolVar(&runRooms, "r", false, "run rooms")
	flag.Parse()

	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.config/nuclasses")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	config, err := pgx.ParseConfig("")
	if err != nil {
		panic(err)
	}
	config.Database = viper.GetString("db.database")
	config.User = viper.GetString("db.user")
	config.Password = viper.GetString("db.password")

	p, err := pgx.ConnectConfig(context.Background(), config)
	if err != nil {
		panic(err)
	}
	defer p.Close(context.Background())

	nu := nuapiclient.NewClient(viper.GetString("nuapi.key"))
	tx, err := p.Begin(context.Background())
	if err != nil {
		panic(err)
	}

	if len(runCourses) > 0 {
		if runCourses == "all" {
			err = allCourses(context.Background(), nu, tx)
			if err != nil {
				_ = tx.Rollback(context.Background())
				panic(err)
			}
		} else {
			err = courses(context.Background(), nu, tx)
			if err != nil {
				_ = tx.Rollback(context.Background())
				panic(err)
			}
		}
	} else {
		if runTerms {
			err = terms(context.Background(), nu, tx)
			if err != nil {
				_ = tx.Rollback(context.Background())
				panic(err)
			}
		}
		if runSchools {
			err = schools(context.Background(), nu, tx)
			if err != nil {
				_ = tx.Rollback(context.Background())
				panic(err)
			}
		}
		if runSubjects {
			err = subjects(context.Background(), nu, tx)
			if err != nil {
				_ = tx.Rollback(context.Background())
				panic(err)
			}
		}
		if runInstructors {
			err = instructors(context.Background(), nu, tx)
			if err != nil {
				_ = tx.Rollback(context.Background())
				panic(err)
			}
		}
		if runBuildings {
			err = buildings(context.Background(), nu, tx)
			if err != nil {
				_ = tx.Rollback(context.Background())
				panic(err)
			}
		}
		if runRooms {
			err = rooms(context.Background(), nu, tx)
			if err != nil {
				_ = tx.Rollback(context.Background())
				panic(err)
			}
		}
	}
	if err := tx.Commit(context.Background()); err != nil {
		panic(err)
	}
}

func allCourses(ctx context.Context, nu *nuapiclient.Client, tx pgx.Tx) error {
	sql := "SELECT id FROM terms"
	rows, err := tx.Query(ctx, sql)
	if err != nil {
		return err
	}
	var terms []int
	for rows.Next() {
		var t int
		if err := rows.Scan(&t); err != nil {
			return err
		}
		terms = append(terms, t)
	}
	rows.Close()

	coursesMap := make(map[int]*nuapiclient.Course)
	for _, t := range terms {
		sql = "SELECT symbol FROM subjects WHERE term_id=$1"
		rows, err := tx.Query(ctx, sql, t)
		if err != nil {
			return err
		}
		var subjects []string
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				return err
			}
			subjects = append(subjects, s)
		}
		rows.Close()

		for _, s := range subjects {
			courses, err := nu.Courses(nuapiclient.CoursesConfig{
				Term:    strconv.Itoa(t),
				Subject: s,
			})
			if err != nil {
				return err
			}
			for _, c := range courses {
				if _, ok := coursesMap[c.ID]; !ok {
					coursesMap[c.ID] = c
				}
			}
		}
	}
	for _, c := range coursesMap {
		sql := "INSERT INTO courses (id, title, term, instructor, subject, catalog_num, section, room, meeting_days, start_time, end_time, seats, topic, component, class_num, course_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)"
		if _, err := tx.Exec(ctx, sql, c.ID, c.Title, c.Term, c.Instructor, c.Subject, c.CatalogNum, c.Section, c.Room, c.MeetingDays, c.StartTime, c.EndTime, c.Seats, c.Topic, c.Component, c.ClassNum, c.CourseID); err != nil {
			return err
		}
	}
	return nil
}

func courses(ctx context.Context, nu *nuapiclient.Client, tx pgx.Tx) error {
	sql := "SELECT id FROM terms WHERE name=$1"
	var term int
	if err := tx.QueryRow(ctx, sql, runCourses).Scan(&term); err != nil {
		return err
	}

	sql = "SELECT symbol FROM subjects WHERE term_id=$1"
	rows, err := tx.Query(ctx, sql, term)
	if err != nil {
		return err
	}
	defer rows.Close()
	var subjects []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return err
		}
		subjects = append(subjects, s)
	}

	coursesMap := make(map[int]*nuapiclient.Course)
	for _, s := range subjects {
		courses, err := nu.Courses(nuapiclient.CoursesConfig{
			Term:    strconv.Itoa(term),
			Subject: s,
		})
		if err != nil {
			return err
		}
		for _, c := range courses {
			if _, ok := coursesMap[c.ID]; !ok {
				coursesMap[c.ID] = c
			}
		}
	}
	for _, c := range coursesMap {
		sql := "INSERT INTO courses (id, title, term, instructor, subject, catalog_num, section, room, meeting_days, start_time, end_time, seats, topic, component, class_num, course_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)"
		if _, err := tx.Exec(ctx, sql, c.ID, c.Title, c.Term, c.Instructor, c.Subject, c.CatalogNum, c.Section, c.Room, c.MeetingDays, c.StartTime, c.EndTime, c.Seats, c.Topic, c.Component, c.ClassNum, c.CourseID); err != nil {
			return err
		}
	}
	return nil
}

func terms(ctx context.Context, nu *nuapiclient.Client, tx pgx.Tx) error {
	terms, err := nu.Terms()
	if err != nil {
		return err
	}
	for _, t := range terms {
		sql := "INSERT INTO terms (id, name, start_date, end_date) VALUES ($1, $2, $3, $4)"
		if _, err := tx.Exec(ctx, sql, t.ID, t.Name, t.StartDate, t.EndDate); err != nil {
			return err
		}
	}
	return nil
}

func schools(ctx context.Context, nu *nuapiclient.Client, tx pgx.Tx) error {
	schools, err := nu.Schools()
	if err != nil {
		return err
	}
	for _, s := range schools {
		sql := "INSERT INTO schools (symbol, name) VALUES ($1, $2)"
		if _, err := tx.Exec(ctx, sql, s.Symbol, s.Name); err != nil {
			return err
		}
	}
	return nil
}

func subjects(ctx context.Context, nu *nuapiclient.Client, tx pgx.Tx) error {
	sql := "SELECT id FROM terms"
	rows, err := tx.Query(ctx, sql)
	if err != nil {
		return err
	}
	var terms []int
	for rows.Next() {
		var t int
		if err := rows.Scan(&t); err != nil {
			return err
		}
		terms = append(terms, t)
	}
	rows.Close()

	sql = "SELECT symbol FROM schools"
	rows, err = tx.Query(ctx, sql)
	if err != nil {
		return err
	}
	var schools []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return err
		}
		schools = append(schools, s)
	}
	rows.Close()

	for _, t := range terms {
		for _, s := range schools {
			subjects, err := nu.Subjects(nuapiclient.SubjectsConfig{
				Term:   strconv.Itoa(t),
				School: s,
			})
			if err != nil {
				return err
			}
			for _, su := range subjects {
				sql := "INSERT INTO subjects (symbol, name, term_id, school_symbol) VALUES ($1, $2, $3, $4)"
				if _, err := tx.Exec(context.Background(), sql, su.Symbol, su.Name, t, s); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func instructors(ctx context.Context, nu *nuapiclient.Client, tx pgx.Tx) error {
	sql := "SELECT DISTINCT symbol FROM subjects"
	rows, err := tx.Query(ctx, sql)
	if err != nil {
		return err
	}
	defer rows.Close()
	var subjects []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return err
		}
		subjects = append(subjects, s)
	}

	instructorsMap := make(map[int]*nuapiclient.Instructor)
	instructorSubjectsMap := make(map[string]map[int]bool)
	for _, s := range subjects {
		instructors, err := nu.Instructors(s)
		if err != nil {
			return err
		}
		for _, i := range instructors {
			if _, ok := instructorsMap[i.ID]; !ok {
				instructorsMap[i.ID] = i
			}
			for _, s := range i.Subjects {
				if _, ok := instructorSubjectsMap[s]; !ok {
					instructorSubjectsMap[s] = make(map[int]bool)
				}
				instructorSubjectsMap[s][i.ID] = true
			}
		}
	}
	for _, i := range instructorsMap {
		sql := "INSERT INTO instructors (id, name, bio, address, phone, office_hours) VALUES ($1, $2, $3, $4, $5, $6)"
		if _, err := tx.Exec(ctx, sql, i.ID, i.Name, i.Bio, i.Address, i.Phone, i.OfficeHours); err != nil {
			return err
		}
	}
	for s, is := range instructorSubjectsMap {
		for i := range is {
			sql := "INSERT INTO instructor_subjects (id, symbol) VALUES ($1, $2)"
			if _, err := tx.Exec(ctx, sql, i, s); err != nil {
				return err
			}
		}
	}
	return nil
}

func buildings(ctx context.Context, nu *nuapiclient.Client, tx pgx.Tx) error {
	buildings, err := nu.Buildings(nuapiclient.BuildingsConfig{})
	if err != nil {
		return err
	}
	for _, b := range buildings {
		sql := "INSERT INTO buildings (id, name, lat, lon, nu_maps_link) VALUES ($1, $2, $3, $4, $5)"
		if _, err := tx.Exec(ctx, sql, b.ID, b.Name, b.Lat, b.Lon, b.NUMapsLink); err != nil {
			return err
		}
	}
	return nil
}

func rooms(ctx context.Context, nu *nuapiclient.Client, tx pgx.Tx) error {
	sql := "SELECT id FROM buildings"
	rows, err := tx.Query(ctx, sql)
	if err != nil {
		return err
	}
	defer rows.Close()
	var buildings []int
	for rows.Next() {
		var b int
		if err := rows.Scan(&b); err != nil {
			return err
		}
		buildings = append(buildings, b)
	}
	for _, b := range buildings {
		rooms, err := nu.Rooms(nuapiclient.RoomsConfig{Building: strconv.Itoa(b)})
		if err != nil {
			return err
		}
		for _, r := range rooms {
			sql := "INSERT INTO rooms (id, building_id, name) VALUES ($1, $2, $3)"
			if _, err := tx.Exec(ctx, sql, r.ID, r.BuildingID, r.Name); err != nil {
				return err
			}
		}
	}
	return nil
}
