import "@fontsource/barlow/400.css";
import "@fontsource/barlow/500.css";
import "@fontsource/barlow/600.css";
import "@fontsource/barlow/700.css";
import "@fontsource/barlow-condensed/600.css";
import "@fontsource/barlow-condensed/700.css";
import "./app.css";
import { mount } from "svelte";
import App from "./App.svelte";
import Overlay from "./overlay/Overlay.svelte";

// One bundle, two windows: the main window and the in-game overlay.
const overlay = new URLSearchParams(location.search).get("view") === "overlay";
if (overlay) document.documentElement.classList.add("overlay-window");
mount(overlay ? Overlay : App, { target: document.getElementById("app")! });
