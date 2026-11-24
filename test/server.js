const express = require("express");
const http = require("http");
const socketIo = require("socket.io");
const app = express();
const server = http.createServer(app);
const io = socketIo(server);

let users = [];
let messages = [];

app.use(express.json());

// Mock login endpoint
app.post("/api/login", (req, res) => {
    const { username, password } = req.body;
    if (username === "test" && password === "password") {
        return res.json({ token: "fake-jwt-token" });
    }
    return res.status(401).json({ message: "Invalid login" });
});

// Serve static HTML files
app.use(express.static("public"));

io.on("connection", (socket) => {
    let username = "";

    // User login
    socket.on("login", (data) => {
        username = data.username;
        users.push({ username, socket });
        console.log(`${username} connected`);
    });

    // Send message to everyone
    socket.on("sendMessage", (data) => {
        const msg = { username: data.username, message: data.message };
        messages.push(msg);
        io.emit("message", msg);
    });

    // Private chat
    socket.on("startPrivateChat", (data) => {
        const friend = users.find(user => user.username === data.username);
        if (friend) {
            console.log(`Private chat started with ${data.username}`);
            socket.emit("message", { username: "System", message: `Started private chat with ${data.username}` });
        }
    });

    socket.on("disconnect", () => {
        users = users.filter(user => user.socket !== socket);
        console.log(`${username} disconnected`);
    });
});

server.listen(4000, () => {
    console.log("Server running on http://localhost:4000");
});
