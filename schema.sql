CREATE TABLE terms (
	id INT PRIMARY KEY,
	name TEXT UNIQUE,
	start_date TEXT,
	end_date TEXT
);

CREATE TABLE schools (
	id SERIAL,
	symbol TEXT UNIQUE,
	name TEXT
);

CREATE TABLE subjects (
	id SERIAL,
	symbol TEXT,
	name TEXT,
	term_id INT REFERENCES terms (id),
	school_symbol TEXT REFERENCES schools (symbol),
	UNIQUE (symbol, term_id, school_symbol)
);

CREATE TABLE instructors (
	id INT PRIMARY KEY,
	name TEXT,
	bio TEXT,
	address TEXT,
	phone TEXT,
	office_hours TEXT
);

CREATE TABLE instructor_subjects (
	id INT REFERENCES instructors (id),
	symbol TEXT,
	UNIQUE (id, symbol)
);

CREATE TABLE buildings (
	id INT PRIMARY KEY,
	name TEXT,
	lat DOUBLE PRECISION,
	lon DOUBLE PRECISION,
	nu_maps_link TEXT
);

CREATE TABLE rooms (
	id INT PRIMARY KEY,
	building_id INT REFERENCES buildings (id),
	name TEXT
);

CREATE TABLE courses (
	id int PRIMARY KEY,
	title TEXT,
	term TEXT REFERENCES terms (name),
	instructor TEXT,
	subject TEXT,
	catalog_num TEXT,
	section TEXT,
	room TEXT,
	meeting_days TEXT,
	start_time TEXT,
	end_time TEXT,
	seats INT,
	topic TEXT,
	component TEXT,
	class_num INT,
	course_id INT
);
