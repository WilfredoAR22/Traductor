package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Establece la conexión a la base de datos
	db, err := conectarDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Configura manejadores para los archivos estáticos (HTML, CSS, imágenes, etc.)
	http.Handle("/", http.FileServer(http.Dir("public")))

	// Manejador para el formulario de inicio de sesión
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		// Obtén los datos del formulario
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Verifica las credenciales del usuario en la base de datos
		err := verificarCredenciales(db, username, password)
		if err != nil {
			http.Error(w, "Credenciales incorrectas", http.StatusUnauthorized)
			return
		}

		// Si las credenciales son válidas, redirige al usuario a Nindex.html
		http.Redirect(w, r, "/src/Ingreso/Nindex.html", http.StatusSeeOther)
	})

	// Manejador para el formulario de registro
	http.HandleFunc("/registro", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		// Obtén los datos del formulario
		nombre := r.FormValue("Nombre")
		correo := r.FormValue("Correo_Electronico")
		usuario := r.FormValue("Usuario")
		contrasenia := r.FormValue("Contrasenia")

		// Guarda los datos en la base de datos
		err := guardarUsuario(db, nombre, correo, usuario, contrasenia)
		if err != nil {
			http.Error(w, "Error al registrar usuario", http.StatusInternalServerError)
			return
		}

		// Redirige al usuario a Nindex.html después del registro
		http.Redirect(w, r, "/src/Ingreso/Nindex.html", http.StatusSeeOther)
	})

	// Manejador para la ruta de diccionario_completo.html
	http.HandleFunc("/src/diccionario/diccionario_completo.html", func(w http.ResponseWriter, r *http.Request) {
		// Realizar la consulta a la base de datos para obtener los datos del diccionario
		resultados, err := consultarDiccionarioCompleto(db)
		if err != nil {
			http.Error(w, "Error al obtener datos del diccionario", http.StatusInternalServerError)
			return
		}

		// Generar el HTML con los resultados de la consulta
		generarHTMLDiccionarioCompleto(w, resultados)
	})

	// Manejador para la ruta de registro de palabras
	http.HandleFunc("/guardar_palabra", func(w http.ResponseWriter, r *http.Request) {
		// Verificar si el método de solicitud es POST
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		// Obtener los datos del formulario
		spanish := r.FormValue("spanishForm")
		qeqchi := r.FormValue("qeqchiForm")

		// Guardar los datos en la base de datos
		err := guardarPalabra(db, spanish, qeqchi)
		if err != nil {
			http.Error(w, "Error al guardar la palabra", http.StatusInternalServerError)
			return
		}

		// Redirigir al usuario a alguna página de éxito o mostrar un mensaje de éxito
		http.Redirect(w, r, "/src/registroP/insertar_datos.html", http.StatusSeeOther)
	})

	// Manejador para la consulta de traducción
	http.HandleFunc("/consulta_qeqchi", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		// Obtén la palabra a traducir del formulario
		palabraQeqchi := r.FormValue("palabra_traducir_qeqchi")

		// Realiza la consulta en la base de datos
		palabraEspanol, err := traducirQeqchi(db, palabraQeqchi)
		if err != nil {
			http.Error(w, "Error al traducir la palabra", http.StatusInternalServerError)
			return
		}

		// Construir una respuesta JSON simple con la traducción
		respuesta := fmt.Sprintf(`{"espaniol": "%s"}`, palabraEspanol)

		// Establecer el encabezado Content-Type para indicar que la respuesta es JSON
		w.Header().Set("Content-Type", "application/json")

		// Escribir la respuesta JSON en el cuerpo de la respuesta
		_, err = w.Write([]byte(respuesta))
		if err != nil {
			http.Error(w, "Error al escribir la respuesta JSON", http.StatusInternalServerError)
			return
		}
	})

	// Manejador para la consulta de traducción inversa (español a Q´eqchi´)
	http.HandleFunc("/consulta_spanish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		// Obtén la palabra en español a traducir del formulario
		palabraEspanol := r.FormValue("palabra_traducir_espa")

		// Realiza la consulta en la base de datos
		palabraQeqchi, err := traducirSpanish(db, palabraEspanol)
		if err != nil {
			http.Error(w, "Error al traducir la palabra", http.StatusInternalServerError)
			return
		}

		// Construir una respuesta JSON simple con la traducción
		respuesta := fmt.Sprintf(`{"qeqchi": "%s"}`, palabraQeqchi)

		// Establecer el encabezado Content-Type para indicar que la respuesta es JSON
		w.Header().Set("Content-Type", "application/json")

		// Escribir la respuesta JSON en el cuerpo de la respuesta
		_, err = w.Write([]byte(respuesta))
		if err != nil {
			http.Error(w, "Error al escribir la respuesta JSON", http.StatusInternalServerError)
			return
		}
	})

	// Inicia el servidor en el puerto 8080
	http.ListenAndServe(":8080", nil)
}

func conectarDB() (*sql.DB, error) {
	// Configura la conexión a la base de datos MySQL
	db, err := sql.Open("mysql", "root:4418@tcp(127.0.0.1:3306)/bdpftraductor")
	if err != nil {
		return nil, err
	}

	// Verifica la conexión a la base de datos
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Conexión exitosa a la base de datos MySQL")

	return db, nil
}

// funcion para registrar usuarios
func guardarUsuario(db *sql.DB, nombre, correo, usuario, contrasenia string) error {
	// Consulta SQL para insertar un nuevo usuario en la base de datos
	query := "INSERT INTO tb_Usuario (Nombre, Correo_Electronico, Usuario, Contrasenia) VALUES (?, ?, ?, ?)"
	_, err := db.Exec(query, nombre, correo, usuario, contrasenia)
	if err != nil {
		return err
	}

	fmt.Println("Usuario registrado exitosamente")

	return nil
}

// funcion para validacion de usuarios
func verificarCredenciales(db *sql.DB, username, password string) error {
	// Consulta SQL para verificar las credenciales del usuario
	query := "SELECT COUNT(*) FROM tb_Usuario WHERE Usuario = ? AND Contrasenia = ?"
	var count int
	err := db.QueryRow(query, username, password).Scan(&count)
	if err != nil {
		return err
	}

	// Si no se encuentra ningún usuario con las credenciales proporcionadas, devuelve un error
	if count == 0 {
		return fmt.Errorf("credenciales incorrectas")
	}

	return nil
}

// función para consultar el diccionario completo desde la base de datos
func consultarDiccionarioCompleto(db *sql.DB) ([]DiccionarioEntry, error) {
	// Consulta SQL para obtener los datos del diccionario
	query := "SELECT id_Palabra, espaniol, qeqchi FROM tb_traductor"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Estructura para almacenar los resultados de la consulta
	var resultados []DiccionarioEntry

	// Recorrer los resultados de la consulta y almacenarlos en la estructura
	for rows.Next() {
		var entry DiccionarioEntry
		err := rows.Scan(&entry.ID, &entry.Espaniol, &entry.Qeqchi)
		if err != nil {
			return nil, err
		}
		resultados = append(resultados, entry)
	}

	return resultados, nil
}

func generarHTMLDiccionarioCompleto(w http.ResponseWriter, resultados []DiccionarioEntry) {
	// Parsear el archivo HTML de la plantilla
	tmpl, err := template.ParseFiles("public/src/diccionario/diccionario_completo.html")
	if err != nil {
		http.Error(w, "Error al cargar la plantilla HTML", http.StatusInternalServerError)
		return
	}

	// Ejecutar la plantilla HTML con los resultados de la consulta
	err = tmpl.Execute(w, resultados)
	if err != nil {
		http.Error(w, "Error al renderizar la plantilla HTML", http.StatusInternalServerError)
		return
	}
}

// Estructura para almacenar las entradas del diccionario
type DiccionarioEntry struct {
	ID       string
	Espaniol string
	Qeqchi   string
}

func guardarPalabra(db *sql.DB, spanish, qeqchi string) error {
	// Ejecutar la consulta SQL para insertar la palabra en la base de datos
	_, err := db.Exec("INSERT INTO tb_traductor (espaniol, qeqchi) VALUES (?, ?)", spanish, qeqchi)
	if err != nil {
		fmt.Println("Error al ejecutar la consulta SQL:", err)
		return err
	}

	fmt.Println("Palabra guardada correctamente")

	return nil
}

func traducirQeqchi(db *sql.DB, palabraQeqchi string) (string, error) {
	// Realizar la consulta en la base de datos para obtener la traducción
	var palabraEspanol string
	err := db.QueryRow("SELECT espaniol FROM tb_traductor WHERE qeqchi = ?", palabraQeqchi).Scan(&palabraEspanol)
	if err != nil {
		return "", err
	}

	return palabraEspanol, nil
}

// función para traducir de español a Q´eqchi´
func traducirSpanish(db *sql.DB, palabraEspanol string) (string, error) {
	// Realizar la consulta en la base de datos para obtener la traducción
	var palabraQeqchi string
	err := db.QueryRow("SELECT qeqchi FROM tb_traductor WHERE espaniol = ?", palabraEspanol).Scan(&palabraQeqchi)
	if err != nil {
		return "", err
	}

	return palabraQeqchi, nil
}
