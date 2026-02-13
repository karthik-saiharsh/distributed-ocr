# README

## Project Setup

#### Step 1: Installing Wails:
- To Install Wails, you must have Go installed first. (https://go.dev/doc/install)
- After you install Go, to install visit (https://wails.io/docs/gettingstarted/installation) instructions are available for windows, mac and linux

> Note: A few additional steps are required for setting up wails on linux, this is also mentioned in the installation link.

#### Step 2: Development
- The Project has a `/frontend` directory which has the source for all the frontend, with `react` and `vite` project with `npm`.
- For developing frontend, `cd` into `/frontend` and run `npm run dev` (you have to install dependencies first (`npm install`))
- For those developing the backend, the `app.go` is the file to expose all the functions that will be available in the frontend.
- To run the backend you have to run `wails dev` in the root of this project.
- On linux alone you have to run it with `wails dev -tags webkit2_41`.
- `main.go` has the setup and details of the App. There mostly won't be any need to edit this file.
- Lastly, any new feature you plan to add, make a new branch and send a PR, do not commit to main directly.
- Anytime you want to change the frontend, install new dependencies or tools, or run any `npm` related command, do it only in `/frontend` directory.