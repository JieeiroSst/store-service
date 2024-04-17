import "./App.css";
import React, { useState, useEffect } from 'react';

function App() {
  const [data, setData] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const fetchData = async () => {
    setIsLoading(true);
    try {
      const response = await fetch('http://localhost:8080/video?live-stream-id=li6bCcEXlJkf9kHpxXfUgdUG');
      if (!response.ok) {
        throw new Error(`API request failed with status ${response.status}`);
      }
      const jsonData = await response.json();
      setData(jsonData);
    } catch (err) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  return (
    <div id="bg_container">
      {data && (
        <ul>
          <iframe
            id="bg"
            src={data.assets.player}
            frameborder="0"
          ></iframe>
        </ul>
      )}
    </div>
  );
}

export default App;
