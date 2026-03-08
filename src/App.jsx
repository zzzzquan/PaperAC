import { useState, useRef, useCallback, useEffect } from "react";
import LandingPage from "./components/LandingPage";
import UploadPage from "./components/UploadPage";

export default function App() {
  const [page, setPage] = useState("landing");
  const [animating, setAnimating] = useState(false);
  const [animClass, setAnimClass] = useState("");
  const timeoutRef = useRef(null);

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const navigate = useCallback((target) => {
    if (animating || target === page) return;
    setAnimating(true);
    setAnimClass("page-fade-out");
    timeoutRef.current = setTimeout(() => {
      setPage(target);
      setAnimClass("page-fade-in");
      timeoutRef.current = setTimeout(() => {
        setAnimClass("");
        setAnimating(false);
      }, 340);
    }, 200);
  }, [animating, page]);

  return (
    <div className={`page-wrapper ${animClass}`}>
      {page === "landing"
        ? <LandingPage onStart={() => navigate("upload")} />
        : <UploadPage onBack={() => navigate("landing")} />}
    </div>
  );
}
