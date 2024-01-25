package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/proullon/ramsql/driver"
)

type Movie struct {
	ID          int64   `json:"id"`
	ImdbID      string  `json:"imdbID"`
	Title       string  `json:"title"`
	Year        int     `json:"year"`
	Rating      float32 `json:"rating"`
	IsSuperHero bool    `json:"isSuperHero"`
}

/*var movies = []Movie{
	{
		ImdbID:      "tt4154796",
		Title:       "Advanger: Endgame",
		Year:        2019,
		Rating:      8.4,
		IsSuperHero: true,
	},
}*/

func getAllMoviesHandler(c echo.Context) error {
	mvs := []Movie{}
	y := c.QueryParam("year")

	if y == "" {
		rows, err := db.Query(`SELECT id, imdbID, title, year, rating, isSuperHero
		FROM goimdb`)
		if err != nil {
			log.Fatal("query error", err)
		}
		defer rows.Close()

		for rows.Next() {
			var m Movie
			if err := rows.Scan(&m.ID, &m.ImdbID, &m.Title, &m.Year, &m.Rating, &m.IsSuperHero); err != nil {
				return c.JSON(http.StatusInternalServerError, "scan:"+err.Error())
			}
			mvs = append(mvs, m)
		}

		if err := rows.Err(); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, mvs)
	}

	year, err := strconv.Atoi(y) // แปลงจากstring เป็นint
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	rows, err := db.Query(`SELECT id, imdbID, title, year, rating, isSuperHero
	FROM goimdb
	WHERE year = ?`, year)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.ImdbID, &m.Title, &m.Year, &m.Rating, &m.IsSuperHero); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		mvs = append(mvs, m)
	}

	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mvs)
}

func getAllMoviesByIdHandler(c echo.Context) error {
	imdbID := c.Param("imdbID")

	row := db.QueryRow(`SELECT id, imdbID, title, year, rating, isSuperHero 
	FROM goimdb WHERE imdbID=?`, imdbID)
	m := Movie{}
	err := row.Scan(&m.ID, &m.ImdbID, &m.Title, &m.Year, &m.Rating, &m.IsSuperHero)
	switch err {
	case nil:
		return c.JSON(http.StatusOK, m)
	case sql.ErrNoRows:
		return c.JSON(http.StatusNotFound, map[string]string{"message!": "not found"})
	default:
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	/*for _, m := range movies {
		if m.ImdbID == imdbID {
			return c.JSON(http.StatusOK, m)
		}

	}

	return c.JSON(http.StatusNotFound, map[string]string{"message": "not found"})*/
}

func createMoviesHandler(c echo.Context) error { //create ผ่านjson
	m := &Movie{}
	if err := c.Bind(m); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	//movies = append(movies, *m)
	stmt, err := db.Prepare(`
	INSERT INTO goimdb(imdbID,title,year,rating,isSuperHero)
	VALUES (?,?,?,?,?);
	`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer stmt.Close()

	//b := fmt.Sprintf("%v", m.IsSuperHero)
	r, err := stmt.Exec(m.ImdbID, m.Title, m.Year, m.Rating, m.IsSuperHero)
	switch {
	case err == nil:
		id, _ := r.LastInsertId()
		m.ID = id
		return c.JSON(http.StatusCreated, m)
	case err.Error() == "UNIQUE constraint violation":
		return c.JSON(http.StatusConflict, "movie already exists")
	default:
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
}

var db *sql.DB

func conn() {
	var err error
	db, err = sql.Open("ramsql", "goimdb")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	conn()

	createTb := `
	CREATE TABLE IF NOT EXISTS goimdb (
	id INT AUTO_INCREMENT,
	imdbID TEXT NOT NULL UNIQUE,
	title TEXT NOT NULL,
	year INT NOT NULL,
	rating FLOAT NOT NULL,
	isSuperHero BOOLEAN NOT NULL,
	PRIMARY KEY (id)
	);
	`
	if _, err := db.Exec(createTb); err != nil {
		log.Fatal("create table error", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/movies", getAllMoviesHandler)             //part ไปที่ฟังก์ชั่นแรก
	e.GET("/movies/:imdbID", getAllMoviesByIdHandler) //part ไปที่ฟังก์ชั่นที่สอง
	e.POST("/movies", createMoviesHandler)

	port := "2565"
	log.Println("start ... port:", port)

	log.Fatal(e.Start(":" + port))

}
