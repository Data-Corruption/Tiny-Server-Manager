<div align="center">
  <img src="./src/public/logo.svg" alt="Logo" width="100" height="100">
  <h1>Tiny Server Manager (TSM)</h1>
  <i>Dedicated server manager with web UI and backup management for Linux.</i>
</div>

## About TSM

Tiny Server Manager (TSM) is a Linux tool designed to automate and simplify the management of video game servers. Built with Golang, it offers a web UI for easy server lifecycle management (restart, update), backup creation, downloading, and restoration. TSM is configured through a simple JSON file, allowing you to specify a server executable path, server save location, and an administrative password for secure access to the web interface.

### Usage Guide

#### Step 1: Prerequisites
TSM expects a dedicated server exe and a singular save folder / file location. Dedicated servers that don't have this structure are not compatible with TSM. So before using TSM, first have your dedicated server installed and make sure you know how to run it manually!

#### Step 2: Download TSM
1. Visit the TSM release page and download the latest version.
2. Choose a location on your computer where you'd like to keep TSM, and extract it there.

#### Step 3: Configure TSM
1. Open the `config.json` file with a text editor. This file tells TSM how to run your game server.
2. You'll need to fill in three details:
   - `"game_exe_path": "example/path/run.sh",` – Set this variable to the path of your game's server executable file.
   - `"game_save_path": "example/path/saves/slot0",` – Set this variable to the path where your active save is.
   - `"admin_password": "example password",` – Set this variable to a strong password for accessing the TSM web UI.
3. Save your changes and close the file.

#### Step 4: Running TSM and Accessing the Web UI
1. Open your terminal and navigate to your TSM folder.
2. Run the `./bin-linux-{your_arch}`. TSM should start your game server and the web UI. TSM will try and keep the game server tied to itself, starting the game server when it starts, and stopping the game server when it stops.
3. TSM will display the address of the web UI in the terminal (e.g., `http://localhost:80`). Open this address in a web browser to manage your server and backups.

#### Privileged ports!

For using ports below 1024 (privileged ports) you'll need to give TSM permission to do so. You can accomplish this by running the following:
```shell
sudo setcap 'cap_net_bind_service=+ep' /path/to/tsm
```

#### Accessing the web UI outside your local network

If you would like to access the web UI from outside your local network you'll need to forward the port it's on. After forwarding the port you'll then be able to access it via `http://your_public_ip:port`

#### Encryption (HTTPS)

If you want the connection to be encrypted you'll need to own a domain, direct it to your ip, and use something like https://certbot.eff.org/instructions?ws=other&os=ubuntufocal to generate tls cert and key files, then copy them to a folder ./tsm has access to, then set those paths in the config.
