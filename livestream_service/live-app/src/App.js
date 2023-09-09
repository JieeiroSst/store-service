import React, { useState, useEffect } from "react";
import "./App.css";
import { w3cwebsocket as WebSocket } from "websocket";

import { BrowserRouter as Router, Route, Switch,Link } from 'react-router-dom';

const client = new WebSocket("ws://localhost:8080/stream");

function App() {
  return (
    <Router>
      <div>
        <ul>
          <li>
            <Link to="/">Home</Link>
          </li>
          <li>
            <Link to="/about">About</Link>
          </li>
          <li>
            <Link to="/dashboard">Dashboard</Link>
          </li>
          <li>
            <Link to="/chat">Chat</Link>
          </li>
        </ul>

        <hr />

        <Switch>
          <Route exact path="/">
            <Home />
          </Route>
          <Route path="/about">
            <About />
          </Route>
          <Route path="/dashboard">
            <Dashboard />
          </Route>
          <Route path="/chat">
            <Chat />
          </Route>
        </Switch>
      </div>
    </Router>
  );
}

function Chat() {
  const [messages, setMessages] = useState([]);

  useEffect(() => {
    client.onopen = () => {
      console.log("WebSocket Client Connected");
    };
    client.onmessage = (message) => {
      setMessages((prevMessages) => [...prevMessages, message.data]);
    };
    client.onclose = () => {
      console.log("WebSocket Client Disconnected");
    };
  }, []);

  const sendMessage = (event) => {
    event.preventDefault();
    const message = event.target.elements.message.value;
    client.send(message);
  };

  return (
    <div className="App">
      <h1>Streaming Application</h1>
      <form onSubmit={sendMessage}>
        <input type="text" name="message" />
        <button type="submit">Send</button>
      </form>
      <ul>
        {messages.map((message, index) => (
          <li key={index}>{message}</li>
        ))}
      </ul>
    </div>
  );
}

function Home() {
  return (
    <div>
      <h2>Home</h2>
    </div>
  );
}

function About() {
  return (
    <div>
      <h2>About</h2>
    </div>
  );
}

function Dashboard() {
  return (
    <div>
      <h2>Dashboard</h2>
    </div>
  );
}



export default App;