import express, { Request, Response } from "express";
import {
  getDevices,
  getPolicies,
  getDevice,
  getPolicy,
  patchDevice,
  getInfo,
  resurectDevice,
  deletePolicy,
  postPolicy,
} from "./mock";

const app = express();
const port = 8081;

app.use(function (req, res, next) {
  res.header("Access-Control-Allow-Origin", "http://localhost:8080"); // update to match the domain/port you will make the request from or use *
  res.header(
    "Access-Control-Allow-Headers",
    "Origin, X-Requested-With, Content-Type, Accept"
  );
  next();
});

app.get("/", (_: Request, res: Response) => {
  res.sendFile(__dirname + "/index.html");
});

// Ping: GET /v2/info => Info
app.get('/v2/info', getInfo)

// Fetch-Devices: GET /v2/devices/:id => Device
app.get('/v2/devices/:id', getDevice)

// Fetch-Devices: GET /v2/devices?iterator=i => Device[]
app.get('/v2/devices', getDevices)

// Rename-Device: PATCH /v2/devices/:id (Device.name) => Device
// Assign-Tag: PATCH /v2/devices/:id (Device.attributes) => Device
// Retire-Device: PATCH /v2/devices/:id (Device.state) => Device
app.patch('/v2/devices/:id', patchDevice)

// Resurect-Device: POST /v2/devices/:id/resurect => Device
app.post('/v2/devices/:id/resurect', resurectDevice)

// Fetch-Policies: GET /v2/policies/:id => Policy
app.get('/v2/policies/:id', getPolicy)

// Fetch-Policies: GET /v2/policies?iterator=i => Policy[]
app.get('/v2/policies', getPolicies)

// Accept-Changes: POST /v2/policies => Policy
// Schedule-Update: POST /v2/policies => Policy
app.post('/v2/policies', postPolicy)

// Abort-Update: DELETE /v2/policies/:id => Policy
app.delete('/v2/policies/:id', deletePolicy)

app.listen(port, () => {
  // tslint:disable-next-line:no-console
  console.log(`server started at http://localhost:${ port }`);
});
