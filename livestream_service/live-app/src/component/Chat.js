import React, { useState, useEffect } from "react";
import { w3cwebsocket as WebSocket } from "websocket";

const client = new WebSocket("ws://localhost:8080/stream");

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
      // <div className="App">

      // </div>
      <section class="bg-white dark:bg-gray-900">
        <div class="container px-6 py-12 mx-auto">
          <div class="lg:flex lg:items-center lg:-mx-6">

            <div
                      class="w-full px-8 py-10 mx-auto overflow-hidden bg-white rounded-lg shadow-2xl dark:bg-gray-900 lg:max-w-xl shadow-gray-300/50 dark:shadow-black/50">

            <form class="mt-6" onSubmit={sendMessage}>
              
              <div class="flex-1">
                <label class="block mb-2 text-sm text-gray-600 dark:text-gray-200">Message</label>
                <textarea type="text" name="message" class="block w-full px-5 py-3 mt-2 text-gray-700 w-full px-5 py-3 mt-2 text-gray-700 border border-gray-200" />
              </div>

              <button type="submit" class="w-full px-6 py-3 mt-6 text-sm font-medium">Send</button>

            </form>

            <ul>
              {messages.map((message, index) => (
                <li key={index}>{message}</li>
              ))}
            </ul>          

            </div>
          </div>
        </div>
       </section>
    );
  }
  

export default Chat;