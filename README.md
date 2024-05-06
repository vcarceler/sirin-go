# Сирин / Sirin

Sirin es un servicio web permite ejecutar `playbooks` de `Ansible` cuando los ordenadores lo solicitan evitando ejecuciones repetidas durante un periodo de tiempo `timeout`. Se trata de una reimplementación en Go de la [anterior versión implementada como una aplicación python/django](https://github.com/vcarceler/sirin).

Está pensado para que cada ordenador haga una petición a Sirin (normalmente en el momento del arranque) indicando el `playbook` que solicita. Sirin identifica a los equipos por la `ip` desde la que realizan la petición. Registrará la petición si no hay ninguna petición para dicho equipo o si la petición registrada es anterior al `timeout` indicado, en tal caso se sobreescribe la petición registrando el momento actual. Sirin guardará en memoria (nunca guarda los datos en disco) cada solicitud registrando:

 * ADDRESS: Dirección `ip` de ordenador que ha realizado la petición. Sirve para identificar a los equipos. Únicamente se guarda la última petición de cada equipo.
 * PLAYBOOK: `playbook` de `ansible` solicitado. Es una cadena arbitraria pero en el instituto utilizamos directamente el nombre del playbook que hay que aplicar al equipo.
 * TIME: Fecha y hora en la que se registró la petición.
 * PENDING: `true` para solicitudes que se acaban de registrar o se han actualizado porque se han vuelto a registrar pasado su `timeout`. `false` para peticiones registradas cuya dirección `ip` ha sido retornada al consultar `/gethosts/`.

El equipo encargado de ejecutar los `playbooks` de `ansible`(en nuestro instituto [Baba-yaga](https://elpuig.xeill.net/Members/vcarceler/articulos/baba-yaga-renueva-su-hogar)) puede consultar periódicamente (en nuestro instituto cada 5 minutos) la lista de equipos que se han registrado para un determinado `playbook`. Sirin devolverá la lista de direcciones IP (separadas por `,`) para que sea posible lanzar `ansible-playbook` utilizando el parámetro `--limit=` de manera que la ejecución del `playbook` se limite a los equipos registrados.

Cuando Sirin devuelve una lista de equipos entiende que esas peticiones ya han sido procesadas y las elimina.

# Instalación de Sirin

Será suficiente con copiar el ejecutable en algún directorio.

Es posible compilar Sirin clonando el repositorio y ejecutando:

~~~
cd siring-go
go build
~~~

Lo que producirá el elecutable `sirin-go`.

# Ejecución de Sirin

Será posible ejecutar Sirin utilizando los parámetros por defecto:

~~~
vcarceler@sputnik:~/dev/sirin-go$ ./sirin-go 
2024/05/06 16:25:03 sirin -address 0.0.0.0 -port 8080 -secret SIRIN -timeout 23h
~~~

En este caso Sirin:

 * Se conectará a todas las interfaces de red (`0.0.0.0`).
 * Atenderá en el puerto `8080`.
 * Se utilizará como `secret` la cadena `SIRIN`.
 * Se utilizará un `timeout` de `23h`.

 Pero se podrá indicar un valor adecuado para cualquiera de estos parámetros:

 ~~~
 vcarceler@sputnik:~/dev/sirin-go$ ./sirin-go --help
Usage of ./sirin-go:
  -address string
    	Dirección para recibir peticiones (default "0.0.0.0")
  -port int
    	Puerto (default 8080)
  -secret string
    	Token secreto (default "SIRIN")
  -timeout string
    	Tiempo antes de registrar una nueva petición (default "23h")
vcarceler@sputnik:~/dev/sirin-go$
 ~~~

El parámetro `secret` permite especificar la cadena que se utilizará en las peticiones `/gethosts/<secret>/<playbook>` para obtener la lista de equipos que se deben incluir en la ejecución del `playbook`.

El `timeout` se puede especificar de cualquiera de las formas aceptadas por la función [time.ParseDuration()](https://pkg.go.dev/time#ParseDuration).

Durante el funcionamiento se irán registrando las solicitudes recibidas:

~~~
vcarceler@sputnik:~/dev/sirin-go$ ./sirin-go -secret AsÑlkYh -timeout 30s
2024/05/06 16:41:35 sirin -address 0.0.0.0 -port 8080 -secret AsÑlkYh -timeout 30s
2024/05/06 16:41:46 /register/ playbook=playbook1 addr=127.0.0.1 port=40402 newrequest=true
2024/05/06 16:41:51 /register/ playbook=playbook1 addr=127.0.0.1 port=58372 newrequest=false elapsed=5.110746816s timeout=30s discarded
2024/05/06 16:42:18 /register/ playbook=playbook1 addr=127.0.0.1 port=49332 newrequest=false elapsed=32.090897384s timeout=30s updated
2024/05/06 16:42:51 /listpendingrequests/ remoteaddress=127.0.0.1:36738 count=1
2024/05/06 16:43:09 /getnumberofrequests/ playbook= addr=127.0.0.1 port=51788 count=0
2024/05/06 16:43:35 /getnumberofrequests/ playbook=playbook1 addr=127.0.0.1 port=39080 count=1
2024/05/06 16:44:09 /gethosts/aeiou/playbook1 playbook=playbook1 remoteaddress=127.0.0.1:36038 error='bad secret'
2024/05/06 16:44:39 /gethosts/ playbook=playbook1 addr=127.0.0.1 port=57696 count=1 hosts=127.0.0.1,
~~~

# Registro de un equipo

Un equipo podrá registrase en Sirin haciendo una petición con `wget`.

~~~
wget http://127.0.0.1:8080/register/playbook.yml
~~~

# Consulta de equipos registrados

Será posible obtener un listado de los equipos registrados accediendo a `http://127.0.0.1:8080/listpendingrequests/`.

~~~
2024-05-06 16:52:51.83 10.73.138.45 playbook2.yml
2024-05-06 16:42:18.15 127.0.0.1 playbook1
2024-05-06 16:52:24.69 10.73.138.177 playbook1.yml
~~~

# Número de solicitandes de un playbook

Acceder a `http://127.0.0.1:8080/getnumberofrequests/playbook1.yml` retornará el número de solicitudes del `playbook` indicado.

# Obtención de los hosts para un playbook

Al acceder a `http://127.0.0.1:8080/gethosts/<SECRET>/<PLAYBOOK>` se obtendrá el listado de equipos que deben incluirse en la ejecución del playbook. Además Sirin dará por procesadas las solicitudes de estos equipos y las marcará como procesadas.

Las solicitudes procesadas (estado `pending=false`) pueden volver a estar pendientes (`pending=true`) si se vuelven a recibir una vez que ha transcurrido su `timeout`.

## Built with ❤️

* [Go](https://go.dev/) - Build simple, secure, scalable systems with Go.
* [GNU/Linux](https://es.wikipedia.org/wiki/GNU/Linux) - Un sistema operativo libre.

## Authors

* Victor Carceler

## License

This project is licensed under the GNU General Public License v3.0 - see the [COPYING](COPYING) file for details.