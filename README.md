#Pretty-URL-Local-Hosting-Basic-JSON-DB

naqiu/Pretty-URL-Local-Hosting-Basic-JSON-DB

A professional local multi-site web server featuring XML Configuration, Pretty URLs, and an integrated JSON Database. This tool allows you to host multiple local domains (Virtual Hosts) from a single instance with automatic system hosts file synchronization.

##🚀 Overview

This application uses a Config.xml file to map local domains to specific paths on your machine. It combines professional-grade multi-site hosting with a lightweight JSON flat-file database system, making it ideal for developers who need to test domain-based configurations locally without manual DNS management.

##✨ Key Features

Multi-Site Hosting: Map different local domains (Virtual Hosts) to specific files or folders via a central XML config.

XML Configuration: Manage ports, domains, and site paths in a clean, structured Config.xml.

Pretty URLs: Professional, local domain-based Virtual Hosting enabled by automatically editing and syncing the system hosts file.

Auto-Open: Automatically launches your browser to specific sites on startup.

JSON Database: * Store data in db/*.json files.

Access via API: GET / POST to /api/db/<filename>.

Example: /api/db/guestbook reads or writes to db/guestbook.json.

##🛠️ Configuration (Config.xml)

The server behavior is defined in Config.xml. You can specify a global port and multiple individual site mappings:

<Config>
    <Port>9000</Port>
    <Sites>
        <Site>
            <Domain>guestbook.local</Domain>
            <Path>example.html</Path>
            <AutoOpen>true</AutoOpen>
        </Site>
        <Site>
            <Domain>docs.local</Domain>
            <Path>./docs</Path>
            <!-- AutoOpen defaults to false if missing -->
        </Site>
    </Sites>
</Config>


##📂 JSON Database API

The built-in database allows you to interact with local JSON files via HTTP requests. All data files are located in the db/ directory.

API Routing Logic

Endpoint: /api/db/<filename>

Storage Path: db/<filename>.json

Method

Request URL

Action

GET

http://guestbook.local:9000/api/db/messages

Returns content of db/messages.json

POST

http://guestbook.local:9000/api/db/messages

Saves/Updates db/messages.json with JSON body


##📄 License

Distributed under the MIT License.
