import {Routes, Route} from "react-router-dom";
import About from "./component/About";
import Blog from "./component/Blog";
import Contact from "./component/Contact";
import Home from "./component/Home";
import Navbar from "./component/Navbar";
import Services from "./component/Services";
import Team from "./component/Team";
import Chat from "./component/Chat";
import User from "./component/User";

function App() {
  return (
    <>
    <Navbar />

    <Routes >
      <Route path="/" element={ <Home /> } />
      <Route path="/blog" element={ <Blog /> } />
      <Route path="/team" element={ <Team /> } />
      <Route path="/about" element={ <About /> } />
      <Route path="/services" element={ <Services /> } />
      <Route path="/contact" element={ <Contact /> } />
      <Route path="/chat" element={ <Chat /> } />
      <Route path="/user" element={ <User /> } />
    </Routes>
    </>
  );
}

export default App;
