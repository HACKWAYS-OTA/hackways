"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import { Icon } from "@/components/ui/Icon";
import "./pass-scanner.css";

export function PassScanner({
  disabled,
  onScan,
}: {
  disabled: boolean;
  onScan: (code: string) => Promise<void>;
}) {
  const video = useRef<HTMLVideoElement>(null),
    canvas = useRef<HTMLCanvasElement>(null),
    upload = useRef<HTMLInputElement>(null);
  const stream = useRef<MediaStream | null>(null),
    frame = useRef(0),
    generation = useRef(0),
    mounted = useRef(true),
    reading = useRef(false);
  const [active, setActive] = useState(false),
    [opening, setOpening] = useState(false),
    [error, setError] = useState(""),
    [canSwitch, setCanSwitch] = useState(false),
    [facing, setFacing] = useState<"environment" | "user">("environment");
  const stop = useCallback(() => {
    generation.current++;
    cancelAnimationFrame(frame.current);
    stream.current?.getTracks().forEach((track) => track.stop());
    stream.current = null;
    if (video.current) video.current.srcObject = null;
    if (mounted.current) {
      setActive(false);
      setOpening(false);
    }
  }, []);
  useEffect(() => {
    mounted.current = true;
    const hide = () => {
      if (document.hidden) stop();
    };
    document.addEventListener("visibilitychange", hide);
    return () => {
      mounted.current = false;
      stop();
      document.removeEventListener("visibilitychange", hide);
    };
  }, [stop]);
  useEffect(() => {
    if (disabled) stop();
  }, [disabled, stop]);
  async function start(direction = facing) {
    if (disabled || opening || reading.current) return;
    stop();
    const request = generation.current;
    setOpening(true);
    setError("");
    try {
      if (!navigator.mediaDevices?.getUserMedia)
        throw new Error(
          "Camera access requires HTTPS and a supported browser. You can choose a QR image below.",
        );
      const camera = await navigator.mediaDevices.getUserMedia({
        audio: false,
        video: {
          facingMode: { ideal: direction },
          width: { ideal: 1280 },
          height: { ideal: 720 },
        },
      });
      if (generation.current !== request || !mounted.current) {
        camera.getTracks().forEach((track) => track.stop());
        return;
      }
      stream.current = camera;
      camera.getVideoTracks().forEach((track) =>
        track.addEventListener("ended", () => {
          if (generation.current === request && mounted.current) {
            stop();
            setError(
              "The camera stopped. Start the scanner again to continue.",
            );
          }
        }),
      );
      void navigator.mediaDevices
        .enumerateDevices()
        .then((devices) => {
          if (generation.current === request && mounted.current)
            setCanSwitch(
              devices.filter((device) => device.kind === "videoinput").length >
                1,
            );
        })
        .catch(() => {
          if (mounted.current) setCanSwitch(false);
        });
      if (!video.current) {
        stop();
        return;
      }
      video.current.srcObject = camera;
      await video.current.play();
      const decode = (await import("jsqr")).default;
      if (generation.current !== request || !mounted.current) return;
      setActive(true);
      setOpening(false);
      let last = 0;
      const tick = (time: number) => {
        if (generation.current !== request || !mounted.current) return;
        const source = video.current,
          surface = canvas.current;
        if (source && surface && source.readyState >= 2 && time - last >= 180) {
          last = time;
          surface.width = Math.min(800, source.videoWidth);
          surface.height = Math.round(
            (source.videoHeight * surface.width) / source.videoWidth,
          );
          const context = surface.getContext("2d", {
            willReadFrequently: true,
          });
          if (context && surface.width && surface.height) {
            context.drawImage(source, 0, 0, surface.width, surface.height);
            const pixels = context.getImageData(
                0,
                0,
                surface.width,
                surface.height,
              ),
              result = decode(pixels.data, pixels.width, pixels.height, {
                inversionAttempts: "dontInvert",
              });
            if (result?.data) {
              stop();
              reading.current = true;
              void onScan(result.data).finally(() => {
                reading.current = false;
              });
              return;
            }
          }
        }
        frame.current = requestAnimationFrame(tick);
      };
      frame.current = requestAnimationFrame(tick);
    } catch (err) {
      if (generation.current !== request || !mounted.current) return;
      stop();
      setError(
        err instanceof DOMException && err.name === "NotAllowedError"
          ? "Camera permission was denied. Allow the camera in your browser, or choose a QR image."
          : err instanceof DOMException && err.name === "NotFoundError"
            ? "No camera was found. Choose a QR image or use manual check-in."
            : err instanceof Error
              ? err.message
              : "The camera could not be opened. Try again.",
      );
    }
  }
  async function readImage(file: File | undefined) {
    if (!file || disabled || reading.current) return;
    stop();
    setError("");
    if (
      !["image/png", "image/jpeg", "image/webp"].includes(file.type) ||
      file.size > 10 * 1024 * 1024
    ) {
      setError("Choose a PNG, JPEG or WebP QR image under 10 MB.");
      return;
    }
    const request = generation.current;
    reading.current = true;
    setOpening(true);
    try {
      const image = await createImageBitmap(file, { resizeWidth: 1200 }),
        surface = canvas.current;
      if (!surface || generation.current !== request || !mounted.current) {
        image.close();
        return;
      }
      surface.width = image.width;
      surface.height = image.height;
      const context = surface.getContext("2d", { willReadFrequently: true });
      if (!context) {
        image.close();
        throw new Error("Your browser could not read the image.");
      }
      context.drawImage(image, 0, 0);
      image.close();
      const pixels = context.getImageData(0, 0, surface.width, surface.height),
        decode = (await import("jsqr")).default;
      if (generation.current !== request || !mounted.current) return;
      const result = decode(pixels.data, pixels.width, pixels.height);
      if (!result?.data)
        throw new Error(
          "No QR code was found. Choose a clearer image with the whole pass visible.",
        );
      await onScan(result.data);
    } catch (err) {
      if (mounted.current && generation.current === request)
        setError(
          err instanceof Error ? err.message : "This image could not be read.",
        );
    } finally {
      reading.current = false;
      if (mounted.current) setOpening(false);
      if (upload.current) upload.current.value = "";
    }
  }
  return (
    <div className="pass-scanner">
      <div className={`pass-scanner-view${active ? " is-active" : ""}`}>
        <video
          ref={video}
          muted
          playsInline
          aria-label="QR pass camera preview"
          hidden={!active}
        />
        {!active && (
          <div className="pass-scanner-placeholder">
            <Icon name="scan" size={42} />
            <p>{opening ? "Opening scanner…" : "Scan a guest’s QR pass"}</p>
            <span>Position the full code inside the frame.</span>
          </div>
        )}
        <span className="pass-scanner-frame" aria-hidden="true" />
      </div>
      <div className="pass-scanner-actions">
        {active ? (
          <>
            <button
              type="button"
              className="app-button app-button-primary"
              onClick={stop}
            >
              Stop camera
            </button>
            {canSwitch && (
              <button
                type="button"
                className="app-button app-button-outline"
                onClick={() => {
                  const next =
                    facing === "environment" ? "user" : "environment";
                  setFacing(next);
                  void start(next);
                }}
              >
                Switch camera
              </button>
            )}
          </>
        ) : (
          <button
            type="button"
            className="app-button app-button-primary"
            disabled={disabled || opening}
            onClick={() => void start()}
          >
            <Icon name="camera" size={18} />
            {opening ? "Opening…" : "Start scanner"}
          </button>
        )}
        <button
          type="button"
          className="app-button app-button-outline"
          disabled={disabled || opening}
          onClick={() => upload.current?.click()}
        >
          Choose QR image
        </button>
        {opening && !active && (
          <button
            type="button"
            className="app-button app-button-outline"
            onClick={stop}
          >
            Cancel
          </button>
        )}
        <input
          ref={upload}
          className="visually-hidden"
          type="file"
          accept="image/png,image/jpeg,image/webp"
          tabIndex={-1}
          aria-label="Choose QR pass image"
          onChange={(event) => void readImage(event.target.files?.[0])}
        />
      </div>
      <canvas ref={canvas} hidden />
      {error && (
        <p className="journey-error" role="alert">
          {error}
        </p>
      )}
      <p className="pass-scanner-help">
        A valid scan checks the guest in automatically. Camera images stay on
        this device.
      </p>
    </div>
  );
}
