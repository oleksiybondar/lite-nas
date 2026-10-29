// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: thumbnail controls keep each user action close to its handler.
// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: the initial builder keeps its reviewable UI sections together.

import { useLanguage } from "@lite-nas/language";
import AddPhotoAlternateIcon from "@mui/icons-material/AddPhotoAlternate";
import DownloadIcon from "@mui/icons-material/Download";
import ShuffleIcon from "@mui/icons-material/Shuffle";
import Alert from "@mui/material/Alert";
import Autocomplete from "@mui/material/Autocomplete";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import FormControlLabel from "@mui/material/FormControlLabel";
import MenuItem from "@mui/material/MenuItem";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Switch from "@mui/material/Switch";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { PointerEvent, ReactElement } from "react";
import { memo, useEffect, useRef, useState } from "react";
import type { GalleryItem } from "./gallery-model";
import { streamProtocol } from "./StreamContext";
import {
  beginThumbnail,
  drawThumbnailLayer,
  loadOriginalThumbnailImage,
  loadThumbnailPreviews,
  renderThumbnail,
  type ThumbnailImage,
  thumbnailExportSize,
} from "./ThumbnailCanvas";
import {
  arrangeThumbnail,
  aspectValue,
  defaultThumbnailComposition,
  fitThumbnailComposition,
  MAX_RANDOM_THUMBNAIL_LAYERS,
  MAX_THUMBNAIL_LAYERS,
  newThumbnailLayer,
  type ThumbnailDocument,
  type ThumbnailLayer,
  thumbnailCompositionSchema,
  thumbnailDocumentSchema,
  thumbnailLayerSchema,
  thumbnailSourceSchema,
} from "./thumbnail-model";

type Template = "scatter" | "ribbon" | "grid" | "featured";
const backgroundLabelKeys: Record<ThumbnailDocument["background"], string> = {
  transparent: "transparent",
  white: "white",
  cream: "cream",
  paper: "paper",
  charcoal: "charcoal",
};
type Drag = {
  id: string;
  action: "move" | "resize" | "rotate" | "composition";
  startX: number;
  startY: number;
  layer: ThumbnailLayer;
  composition: ThumbnailDocument["composition"];
};
const initialDocument: ThumbnailDocument = {
  version: 1,
  aspect: "16:9",
  background: "cream",
  composition: defaultThumbnailComposition,
  layers: [],
};
const imageUrl = (src: string): string =>
  src.startsWith("data:") ? src : (streamProtocol.image(src) ?? "");
const layerEditSchema = thumbnailLayerSchema.omit({ id: true }).partial();

/** Downloads a generated browser file without persisting it to a service. */
const downloadBlob = (blob: Blob, name: string): void => {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = name;
  link.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
};

/** Reads an uploaded image as a portable data URL, preserving PNG alpha. */
const readUpload = (file: File): Promise<string> =>
  new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () =>
      typeof reader.result === "string"
        ? resolve(reader.result)
        : reject(new Error("Image read failed"));
    reader.onerror = () => reject(new Error("Image read failed"));
    reader.readAsDataURL(file);
  });

/** Offers a searchable, bounded list of gallery stills with lazy image previews. */
const GalleryImagePicker = ({
  items,
  label,
  onChoose,
}: {
  items: readonly GalleryItem[];
  label: string;
  onChoose: (item: GalleryItem) => void;
}): ReactElement => {
  const [value, setValue] = useState<GalleryItem | null>(null);
  const [highlighted, setHighlighted] = useState<GalleryItem | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  useEffect(() => {
    const timeout = window.setTimeout(() => setPreview(highlighted?.url ?? null), 220);
    return () => window.clearTimeout(timeout);
  }, [highlighted]);
  return (
    <Autocomplete
      options={items}
      value={value}
      getOptionLabel={(item) => item.filename}
      filterOptions={(options, state) =>
        options
          .filter((item) => item.filename.toLowerCase().includes(state.inputValue.toLowerCase()))
          .slice(0, 12)
      }
      onHighlightChange={(_, item) => setHighlighted(item)}
      onClose={() => {
        setHighlighted(null);
        setPreview(null);
      }}
      onChange={(_, item) => {
        if (item) onChoose(item);
        setValue(null);
      }}
      renderInput={(params) => <TextField {...params} label={label} size="small" />}
      renderOption={(props, item) => (
        <li {...props} key={item.id}>
          {preview === item.url ? (
            <img
              alt=""
              src={imageUrl(item.url)}
              style={{ width: 42, height: 42, objectFit: "cover", marginRight: 10 }}
            />
          ) : (
            <Box sx={{ width: 42, height: 42, bgcolor: "action.hover", mr: 1.25, flexShrink: 0 }} />
          )}
          {item.filename}
        </li>
      )}
    />
  );
};

/** Paints a list thumbnail from an already decoded, bounded preview. */
const LayerThumbnail = memo(({ image }: { image?: ThumbnailImage }): ReactElement => {
  const ref = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    const canvas = ref.current;
    const context = canvas?.getContext("2d");
    if (!canvas || !context) return;
    context.clearRect(0, 0, canvas.width, canvas.height);
    if (!image) return;
    const scale = Math.min(canvas.width / image.width, canvas.height / image.height);
    const width = image.width * scale;
    const height = image.height * scale;
    context.drawImage(
      image.source,
      (canvas.width - width) / 2,
      (canvas.height - height) / 2,
      width,
      height,
    );
  }, [image]);
  return (
    <canvas
      ref={ref}
      width={38}
      height={38}
      style={{ width: 38, height: 38, marginRight: 8, flexShrink: 0 }}
    />
  );
});

/** Lets users type signed and decimal numbers before committing a validated edit. */
const NumericInput = ({
  label,
  value,
  step = 1,
  onCommit,
}: {
  label: string;
  value: number;
  step?: number;
  onCommit: (value: number) => void;
}): ReactElement => {
  const [draft, setDraft] = useState(String(value));
  const [focused, setFocused] = useState(false);
  useEffect(() => {
    if (!focused) setDraft(String(value));
  }, [focused, value]);
  const commit = (): void => {
    const number = Number(draft);
    if (draft.trim() && Number.isFinite(number)) onCommit(number);
    setFocused(false);
    setDraft(String(value));
  };
  return (
    <TextField
      label={label}
      size="small"
      type="number"
      value={draft}
      inputProps={{ step }}
      onFocus={() => setFocused(true)}
      onChange={(event) => setDraft(event.target.value)}
      onBlur={commit}
      onKeyDown={(event) => {
        if (event.key === "Enter") event.currentTarget.blur();
      }}
    />
  );
};

/** Edits a gallery thumbnail and sends its rendered image to the owning workflow. */
export const ThumbnailMaker = ({
  items,
  onSave,
}: {
  items: readonly GalleryItem[];
  onSave: (image: Blob) => Promise<void>;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const gallery = items.filter((item) => item.kind === "image");
  const [documentValue, setDocument] = useState<ThumbnailDocument>(initialDocument);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [template, setTemplate] = useState<Template>("scatter");
  const [moveComposition, setMoveComposition] = useState(false);
  const [randomCount, setRandomCount] = useState(5);
  const [jsonText, setJsonText] = useState("");
  const [message, setMessage] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [images, setImages] = useState<Map<string, ThumbnailImage>>(new Map());
  const previewCacheRef = useRef<Map<string, ThumbnailImage>>(new Map());
  const previewAbortRef = useRef<AbortController | null>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const stageRef = useRef<HTMLDivElement>(null);
  const dragRef = useRef<Drag | null>(null);
  const dragFrameRef = useRef<number | null>(null);
  const pendingDragUpdateRef = useRef<(() => void) | null>(null);
  const sourceSnapshotRef = useRef<Map<string, string>>(new Map());
  const imageRequestRef = useRef(0);
  const selected = documentValue.layers.find((layer) => layer.id === selectedId) ?? null;

  useEffect(() => {
    const nextSources = new Map(documentValue.layers.map((layer) => [layer.id, layer.src]));
    const previousSources = sourceSnapshotRef.current;
    if (
      nextSources.size === previousSources.size &&
      [...nextSources].every(([id, src]) => previousSources.get(id) === src)
    )
      return;
    previewAbortRef.current?.abort();
    const controller = new AbortController();
    previewAbortRef.current = controller;
    sourceSnapshotRef.current = nextSources;
    const requestId = ++imageRequestRef.current;
    const retained = new Set(documentValue.layers.map((layer) => layer.src));
    for (const src of previewCacheRef.current.keys())
      if (!retained.has(src)) previewCacheRef.current.delete(src);
    const refresh = (): void => {
      if (requestId !== imageRequestRef.current) return;
      setImages(
        new Map(
          documentValue.layers.flatMap((layer) => {
            const image = previewCacheRef.current.get(layer.src);
            return image ? [[layer.id, image] as const] : [];
          }),
        ),
      );
    };
    refresh();
    if (documentValue.layers.length === 0) {
      return;
    }
    void loadThumbnailPreviews(documentValue, previewCacheRef.current, refresh, controller.signal)
      .then((loaded) => {
        if (requestId === imageRequestRef.current) {
          setImages(loaded);
          setMessage(
            loaded.size < documentValue.layers.length
              ? getLabel("moments.thumbnail.loadingSome")
              : null,
          );
        }
      })
      .catch(() => {
        if (requestId === imageRequestRef.current)
          setMessage(getLabel("moments.thumbnail.loadFailed"));
      });
  }, [documentValue, getLabel]);

  useEffect(
    () => () => {
      imageRequestRef.current += 1;
      previewAbortRef.current?.abort();
      if (dragFrameRef.current !== null) window.cancelAnimationFrame(dragFrameRef.current);
    },
    [],
  );

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (canvasRef.current) renderThumbnail(canvasRef.current, documentValue, images, 720);
    }, 16);
    return () => window.clearTimeout(timer);
  }, [documentValue, images]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !moveComposition) return;
    const zoom = (event: WheelEvent): void => {
      event.preventDefault();
      setDocument((current) => {
        const scale = Math.max(
          0.25,
          Math.min(4, current.composition.scale * Math.exp(-event.deltaY * 0.001)),
        );
        const result = thumbnailCompositionSchema.safeParse({ ...current.composition, scale });
        return result.success ? { ...current, composition: result.data } : current;
      });
    };
    canvas.addEventListener("wheel", zoom, { passive: false });
    return () => canvas.removeEventListener("wheel", zoom);
  }, [moveComposition]);

  /** Applies a validated edit to the selected layer. */
  const updateLayer = (id: string, changes: Partial<ThumbnailLayer>): void => {
    const validated = layerEditSchema.safeParse(changes);
    if (!validated.success) return;
    setDocument((current) => ({
      ...current,
      layers: current.layers.map((layer) =>
        layer.id === id ? { ...layer, ...validated.data } : layer,
      ),
    }));
  };

  /** Applies bounded edits to the transform shared by all image layers. */
  const updateComposition = (changes: Partial<ThumbnailDocument["composition"]>): void => {
    setDocument((current) => {
      const result = thumbnailCompositionSchema.safeParse({ ...current.composition, ...changes });
      return result.success ? { ...current, composition: result.data } : current;
    });
  };

  /** Repositions an existing set of images and fits its outer corners. */
  const applyTemplate = (): void => {
    setDocument((current) => {
      const layers = arrangeThumbnail(current.layers, template, aspectValue(current.aspect));
      return {
        ...current,
        layers,
        composition: fitThumbnailComposition(layers, current.aspect),
        background: template === "ribbon" ? "transparent" : current.background,
      };
    });
  };

  /** Appends one source while respecting the document's image limit. */
  const addSource = (src: string, name: string): void => {
    if (!thumbnailSourceSchema.safeParse(src).success) {
      setMessage(getLabel("moments.thumbnail.unsupported"));
      return;
    }
    if (documentValue.layers.length >= MAX_THUMBNAIL_LAYERS) {
      setMessage(
        getLabel("moments.thumbnail.maxLayers").replace("{count}", String(MAX_THUMBNAIL_LAYERS)),
      );
      return;
    }
    const layer = newThumbnailLayer(src, name);
    setDocument((current) => ({ ...current, layers: [...current.layers, layer] }));
    setSelectedId(layer.id);
    setMessage(null);
  };

  /** Embeds each accepted upload in the in-memory recipe. */
  const onUpload = async (files: FileList | null): Promise<void> => {
    if (!files) return;
    if (files.length > MAX_THUMBNAIL_LAYERS - documentValue.layers.length) {
      setMessage(
        getLabel("moments.thumbnail.maxLayers").replace("{count}", String(MAX_THUMBNAIL_LAYERS)),
      );
      return;
    }
    const allowed = Array.from(files);
    try {
      const layers: ThumbnailLayer[] = [];
      for (const file of allowed) {
        if (!/^image\/(png|jpeg|webp|gif)$/.test(file.type))
          throw new Error(getLabel("moments.thumbnail.formats"));
        if (file.size > 20 * 1024 * 1024) throw new Error(getLabel("moments.thumbnail.maxUpload"));
        const src = await readUpload(file);
        if (!thumbnailSourceSchema.safeParse(src).success)
          throw new Error(getLabel("moments.thumbnail.validateUpload"));
        layers.push(newThumbnailLayer(src, file.name));
      }
      setDocument((current) => ({
        ...current,
        layers: [...current.layers, ...layers].slice(0, MAX_THUMBNAIL_LAYERS),
      }));
      if (layers[0]) setSelectedId(layers[0].id);
      setMessage(null);
    } catch (error) {
      setMessage(
        error instanceof Error ? error.message : getLabel("moments.thumbnail.uploadFailed"),
      );
    }
  };

  /** Picks up to the random-selection limit of distinct gallery images. */
  const randomize = (): void => {
    if (gallery.length === 0) {
      setMessage(getLabel("moments.thumbnail.notReady"));
      return;
    }
    const count = Math.min(MAX_RANDOM_THUMBNAIL_LAYERS, Math.max(1, randomCount), gallery.length);
    const shuffled = [...gallery];
    for (let index = shuffled.length - 1; index > 0; index -= 1) {
      const randomIndex = Math.floor(Math.random() * (index + 1));
      [shuffled[index], shuffled[randomIndex]] = [shuffled[randomIndex], shuffled[index]];
    }
    shuffled.length = count;
    const layers = arrangeThumbnail(
      shuffled.map((item) => newThumbnailLayer(item.url, item.filename)),
      template,
      aspectValue(documentValue.aspect),
    );
    setDocument((current) => ({
      ...current,
      layers,
      composition: fitThumbnailComposition(layers, current.aspect),
      background: template === "ribbon" ? "transparent" : current.background,
    }));
    setSelectedId(layers[0]?.id ?? null);
    setMessage(null);
  };

  /** Starts pointer movement, resizing, or rotation for a layer. */
  const startDrag = (
    event: PointerEvent<HTMLElement>,
    layer: ThumbnailLayer,
    action: Drag["action"],
  ): void => {
    event.stopPropagation();
    event.currentTarget.setPointerCapture(event.pointerId);
    dragRef.current = {
      id: layer.id,
      action,
      startX: event.clientX,
      startY: event.clientY,
      layer,
      composition: documentValue.composition,
    };
    setSelectedId(layer.id);
  };

  /** Coalesces pointer changes to one editor update per animation frame. */
  const scheduleDragUpdate = (update: () => void): void => {
    pendingDragUpdateRef.current = update;
    if (dragFrameRef.current !== null) return;
    dragFrameRef.current = window.requestAnimationFrame(() => {
      dragFrameRef.current = null;
      pendingDragUpdateRef.current?.();
      pendingDragUpdateRef.current = null;
    });
  };

  /** Applies the final pointer position before releasing an edit gesture. */
  const finishDrag = (): void => {
    if (dragFrameRef.current !== null) window.cancelAnimationFrame(dragFrameRef.current);
    dragFrameRef.current = null;
    pendingDragUpdateRef.current?.();
    pendingDragUpdateRef.current = null;
    dragRef.current = null;
  };

  /** Converts pointer displacement into normalized document coordinates. */
  const moveDrag = (event: PointerEvent<HTMLElement>): void => {
    const drag = dragRef.current;
    const rect = stageRef.current?.getBoundingClientRect();
    if (!drag || !rect) return;
    const dx = ((event.clientX - drag.startX) / rect.width) * 100;
    const dy = ((event.clientY - drag.startY) / rect.height) * 100;
    if (drag.action === "composition") {
      scheduleDragUpdate(() =>
        updateComposition({ x: drag.composition.x + dx, y: drag.composition.y + dy }),
      );
      return;
    }
    const localDx = dx / drag.composition.scale;
    const localDy = dy / drag.composition.scale;
    if (drag.action === "move")
      scheduleDragUpdate(() =>
        updateLayer(drag.id, { x: drag.layer.x + localDx, y: drag.layer.y + localDy }),
      );
    if (drag.action === "resize")
      scheduleDragUpdate(() =>
        updateLayer(drag.id, {
          width: drag.layer.width + localDx,
          height: drag.layer.height + localDy,
        }),
      );
    if (drag.action === "rotate") {
      const centerX =
        rect.left +
        ((50 +
          drag.composition.x +
          drag.composition.scale * (drag.layer.x + drag.layer.width / 2 - 50)) /
          100) *
          rect.width;
      const centerY =
        rect.top +
        ((50 +
          drag.composition.y +
          drag.composition.scale * (drag.layer.y + drag.layer.height / 2 - 50)) /
          100) *
          rect.height;
      const angle =
        (Math.atan2(event.clientY - centerY, event.clientX - centerX) * 180) / Math.PI + 90;
      scheduleDragUpdate(() =>
        updateLayer(drag.id, { angle: Math.max(-180, Math.min(180, angle)) }),
      );
    }
  };

  /** Selects the topmost image under the pointer and starts moving it. */
  const chooseCanvasLayer = (event: PointerEvent<HTMLCanvasElement>): void => {
    const rect = stageRef.current?.getBoundingClientRect();
    if (!rect) return;
    if (moveComposition) {
      if (selected) startDrag(event, selected, "composition");
      return;
    }
    const x =
      50 +
      (((event.clientX - rect.left) / rect.width) * 100 - 50 - documentValue.composition.x) /
        documentValue.composition.scale;
    const y =
      50 +
      (((event.clientY - rect.top) / rect.height) * 100 - 50 - documentValue.composition.y) /
        documentValue.composition.scale;
    const layer = [...documentValue.layers]
      .reverse()
      .find(
        (candidate) =>
          x >= candidate.x &&
          x <= candidate.x + candidate.width &&
          y >= candidate.y &&
          y <= candidate.y + candidate.height,
      );
    if (layer) startDrag(event, layer, "move");
  };

  /** Renders the validated recipe at export resolution for download or storage. */
  const renderImage = async (format: "png" | "jpeg"): Promise<Blob> => {
    const result = thumbnailDocumentSchema.safeParse(documentValue);
    if (!result.success) throw new Error(getLabel("moments.thumbnail.addImageFirst"));
    if (images.size !== result.data.layers.length)
      throw new Error(getLabel("moments.thumbnail.waitPreviews"));
    const [width, height] = thumbnailExportSize(result.data, images);
    const canvas = window.document.createElement("canvas");
    const context = beginThumbnail(canvas, result.data, width, height, format === "jpeg");
    for (const layer of result.data.layers) {
      const original = await loadOriginalThumbnailImage(layer.src);
      drawThumbnailLayer(context, result.data, layer, original, width, height);
      if (original.source instanceof HTMLImageElement) original.source.removeAttribute("src");
    }
    return new Promise((resolve, reject) =>
      canvas.toBlob(
        (blob) => (blob ? resolve(blob) : reject(new Error("Image rendering failed."))),
        `image/${format}`,
        format === "jpeg" ? 0.94 : undefined,
      ),
    );
  };

  /** Downloads a rendered image without altering the gallery cover. */
  const exportImage = async (format: "png" | "jpeg"): Promise<void> => {
    try {
      const blob = await renderImage(format);
      downloadBlob(blob, `moments-thumbnail.${format === "jpeg" ? "jpg" : "png"}`);
    } catch {
      setMessage(getLabel("moments.thumbnail.exportFailed"));
    }
  };

  /** Publishes the rendered PNG as this gallery's saved cover. */
  const saveImage = async (): Promise<void> => {
    setSaving(true);
    try {
      await onSave(await renderImage("png"));
      setMessage(null);
    } catch {
      setMessage(getLabel("moments.thumbnail.saveFailed"));
    } finally {
      setSaving(false);
    }
  };

  /** Replaces the current editor state with a fully validated recipe. */
  const importJson = (): void => {
    try {
      const parsed: unknown = JSON.parse(jsonText);
      const result = thumbnailDocumentSchema.safeParse(parsed);
      if (!result.success) throw new Error(getLabel("moments.thumbnail.jsonFormat"));
      setDocument(result.data);
      setSelectedId(result.data.layers[0]?.id ?? null);
      setMessage(null);
    } catch {
      setMessage(getLabel("moments.thumbnail.invalidJson"));
    }
  };

  /** Downloads the recipe, including any embedded uploads. */
  const exportJson = (): void => {
    const result = thumbnailDocumentSchema.safeParse(documentValue);
    if (!result.success) {
      setMessage(getLabel("moments.thumbnail.exportJsonFirst"));
      return;
    }
    downloadBlob(
      new Blob([JSON.stringify(result.data, null, 2)], { type: "application/json" }),
      "moments-thumbnail.json",
    );
  };

  const numericField = (
    label: string,
    key: "x" | "y" | "width" | "height" | "angle" | "zoom" | "focusX" | "focusY",
    step = 1,
  ): ReactElement | null =>
    selected && (
      <NumericInput
        label={label}
        value={selected[key]}
        step={step}
        onCommit={(value) => updateLayer(selected.id, { [key]: value })}
      />
    );

  const selectedRatio = selected
    ? (selected.width * aspectValue(documentValue.aspect)) / selected.height
    : 0;
  const ratioOptions = ["1:1", "4:3", "3:4", "16:9", "9:16"];
  const selectedRatioPreset =
    ratioOptions.find(
      (option) =>
        Math.abs(aspectValue(option as ThumbnailDocument["aspect"]) - selectedRatio) < 0.02,
    ) ?? "custom";

  /** Changes the selected image frame's ratio without changing its crop zoom. */
  const setImageRatio = (ratio: number): void => {
    if (!selected || !Number.isFinite(ratio) || ratio <= 0) return;
    let width = selected.width;
    let height = (width * aspectValue(documentValue.aspect)) / ratio;
    if (height > 200) {
      width *= 200 / height;
      height = 200;
    }
    updateLayer(selected.id, { width, height });
  };

  return (
    <Stack spacing={3} data-testid="thumbnail-maker">
      <Box>
        <Typography variant="h4">{getLabel("moments.thumbnail.galleryTitle")}</Typography>
        <Typography color="text.secondary">
          {getLabel("moments.thumbnail.arrangeHelp").replace(
            "{count}",
            String(MAX_THUMBNAIL_LAYERS),
          )}
        </Typography>
      </Box>
      {message && (
        <Alert severity="warning" onClose={() => setMessage(null)}>
          {message}
        </Alert>
      )}
      <Stack direction={{ xs: "column", md: "row" }} spacing={3} alignItems="flex-start">
        <Stack spacing={2} sx={{ width: { xs: "100%", md: 330 }, flexShrink: 0 }}>
          <Paper sx={{ p: 2 }}>
            <Stack spacing={2}>
              <Typography variant="h6">
                {getLabel("moments.thumbnail.images")} ({documentValue.layers.length}/
                {MAX_THUMBNAIL_LAYERS})
              </Typography>
              <GalleryImagePicker
                items={gallery}
                label={getLabel("moments.thumbnail.addGalleryPhoto")}
                onChoose={(item) => addSource(item.url, item.filename)}
              />
              <Button component="label" startIcon={<AddPhotoAlternateIcon />} variant="outlined">
                {getLabel("moments.thumbnail.uploadImages")}
                <input
                  hidden
                  multiple
                  accept="image/png,image/jpeg,image/webp,image/gif"
                  type="file"
                  onChange={(event) => {
                    void onUpload(event.target.files);
                    event.target.value = "";
                  }}
                />
              </Button>
              <Stack direction="row" spacing={1}>
                <TextField
                  label={getLabel("moments.thumbnail.randomCount")}
                  size="small"
                  type="number"
                  value={randomCount}
                  inputProps={{ min: 1, max: MAX_RANDOM_THUMBNAIL_LAYERS }}
                  onChange={(event) =>
                    setRandomCount(
                      Math.max(
                        1,
                        Math.min(MAX_RANDOM_THUMBNAIL_LAYERS, Number(event.target.value) || 1),
                      ),
                    )
                  }
                />
                <Button onClick={randomize} startIcon={<ShuffleIcon />} variant="contained">
                  {getLabel("moments.thumbnail.random")}
                </Button>
              </Stack>
              <TextField
                select
                label={getLabel("moments.thumbnail.template")}
                size="small"
                value={template}
                onChange={(event) => setTemplate(event.target.value as Template)}
              >
                <MenuItem value="scatter">{getLabel("moments.thumbnail.templateScatter")}</MenuItem>
                <MenuItem value="ribbon">{getLabel("moments.thumbnail.templateRibbon")}</MenuItem>
                <MenuItem value="grid">{getLabel("moments.thumbnail.templateGrid")}</MenuItem>
                <MenuItem value="featured">
                  {getLabel("moments.thumbnail.templateFeatured")}
                </MenuItem>
              </TextField>
              <Button disabled={documentValue.layers.length === 0} onClick={applyTemplate}>
                {getLabel("moments.thumbnail.applyTemplate")}
              </Button>
              <Stack sx={{ maxHeight: 250, overflowY: "auto" }} spacing={0.5}>
                {documentValue.layers.map((layer) => (
                  <Button
                    key={layer.id}
                    color={layer.id === selectedId ? "primary" : "inherit"}
                    onClick={() => setSelectedId(layer.id)}
                    sx={{ justifyContent: "flex-start", textTransform: "none" }}
                  >
                    <LayerThumbnail image={images.get(layer.id)} />
                    <Box
                      component="span"
                      sx={{ overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}
                    >
                      {layer.name}
                    </Box>
                  </Button>
                ))}
              </Stack>
            </Stack>
          </Paper>
          <Paper sx={{ p: 2 }}>
            <Stack spacing={2}>
              <Typography variant="h6">{getLabel("moments.thumbnail.canvas")}</Typography>
              <TextField
                select
                label={getLabel("moments.thumbnail.aspectRatio")}
                size="small"
                value={documentValue.aspect}
                onChange={(event) =>
                  setDocument((current) => ({
                    ...current,
                    aspect: event.target.value as ThumbnailDocument["aspect"],
                  }))
                }
              >
                {["16:9", "4:3", "1:1", "3:4", "9:16"].map((aspect) => (
                  <MenuItem key={aspect} value={aspect}>
                    {aspect}
                  </MenuItem>
                ))}
              </TextField>
              <TextField
                select
                label={getLabel("moments.thumbnail.background")}
                size="small"
                value={documentValue.background}
                onChange={(event) =>
                  setDocument((current) => ({
                    ...current,
                    background: event.target.value as ThumbnailDocument["background"],
                  }))
                }
              >
                {["transparent", "white", "cream", "paper", "charcoal"].map((background) => (
                  <MenuItem key={background} value={background}>
                    {getLabel(
                      `moments.thumbnail.background${backgroundLabelKeys[background as ThumbnailDocument["background"]][0].toUpperCase()}${backgroundLabelKeys[background as ThumbnailDocument["background"]].slice(1)}`,
                    )}
                  </MenuItem>
                ))}
              </TextField>
              <Typography variant="subtitle1">
                {getLabel("moments.thumbnail.composition")}
              </Typography>
              <Box sx={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 1 }}>
                <NumericInput
                  label={getLabel("moments.thumbnail.xPercent")}
                  value={documentValue.composition.x}
                  onCommit={(value) => updateComposition({ x: value })}
                />
                <NumericInput
                  label={getLabel("moments.thumbnail.yPercent")}
                  value={documentValue.composition.y}
                  onCommit={(value) => updateComposition({ y: value })}
                />
                <NumericInput
                  label={getLabel("moments.thumbnail.scale")}
                  value={documentValue.composition.scale}
                  step={0.05}
                  onCommit={(value) => updateComposition({ scale: value })}
                />
              </Box>
              <Stack direction="row" spacing={1}>
                <Button
                  onClick={() =>
                    updateComposition({
                      scale: Math.max(0.25, documentValue.composition.scale - 0.1),
                    })
                  }
                >
                  {getLabel("moments.thumbnail.zoomOut")}
                </Button>
                <Button
                  onClick={() =>
                    updateComposition({ scale: Math.min(4, documentValue.composition.scale + 0.1) })
                  }
                >
                  {getLabel("moments.thumbnail.zoomIn")}
                </Button>
              </Stack>
              <Button
                disabled={documentValue.layers.length === 0}
                onClick={() =>
                  updateComposition(
                    fitThumbnailComposition(documentValue.layers, documentValue.aspect),
                  )
                }
              >
                {getLabel("moments.thumbnail.fitComposition")}
              </Button>
              <FormControlLabel
                control={
                  <Switch
                    checked={moveComposition}
                    onChange={(_, checked) => setMoveComposition(checked)}
                  />
                }
                label={getLabel("moments.thumbnail.dragComposition")}
              />
            </Stack>
          </Paper>
        </Stack>
        <Stack spacing={2} sx={{ flex: 1, width: "100%", minWidth: 0 }}>
          <Paper sx={{ p: 2, overflow: "hidden" }}>
            <Box
              ref={stageRef}
              sx={{
                width: "100%",
                aspectRatio: String(aspectValue(documentValue.aspect)),
                position: "relative",
                bgcolor: "#777",
                touchAction: "none",
              }}
            >
              <canvas
                ref={canvasRef}
                aria-label={getLabel("moments.thumbnail.preview")}
                onPointerDown={chooseCanvasLayer}
                onPointerMove={moveDrag}
                onPointerUp={finishDrag}
                onPointerCancel={finishDrag}
                style={{ display: "block", width: "100%", height: "100%" }}
              />
              {selected && !moveComposition && (
                <Box
                  sx={{
                    position: "absolute",
                    inset: 0,
                    transform: `translate(${documentValue.composition.x}%, ${documentValue.composition.y}%) scale(${documentValue.composition.scale})`,
                    transformOrigin: "center center",
                    pointerEvents: "none",
                  }}
                >
                  <Box
                    onPointerDown={(event) => startDrag(event, selected, "move")}
                    onPointerMove={moveDrag}
                    onPointerUp={finishDrag}
                    onPointerCancel={finishDrag}
                    sx={{
                      position: "absolute",
                      left: `${selected.x}%`,
                      top: `${selected.y}%`,
                      width: `${selected.width}%`,
                      height: `${selected.height}%`,
                      transform: `rotate(${selected.angle}deg)`,
                      border: "2px solid #59c8ff",
                      cursor: "move",
                      touchAction: "none",
                      pointerEvents: "auto",
                    }}
                  >
                    <Box
                      onPointerDown={(event) => startDrag(event, selected, "rotate")}
                      onPointerMove={moveDrag}
                      onPointerUp={finishDrag}
                      onPointerCancel={finishDrag}
                      aria-label={getLabel("moments.thumbnail.rotate")}
                      sx={{
                        position: "absolute",
                        top: -18,
                        left: "50%",
                        width: 18,
                        height: 18,
                        borderRadius: "50%",
                        bgcolor: "primary.main",
                        cursor: "grab",
                      }}
                    />
                    <Box
                      onPointerDown={(event) => startDrag(event, selected, "resize")}
                      onPointerMove={moveDrag}
                      onPointerUp={finishDrag}
                      onPointerCancel={finishDrag}
                      aria-label={getLabel("moments.thumbnail.resize")}
                      sx={{
                        position: "absolute",
                        bottom: -9,
                        right: -9,
                        width: 18,
                        height: 18,
                        bgcolor: "primary.main",
                        cursor: "nwse-resize",
                      }}
                    />
                  </Box>
                </Box>
              )}
            </Box>
          </Paper>
          {selected && (
            <Paper sx={{ p: 2 }}>
              <Stack spacing={2}>
                <Typography variant="h6">
                  {getLabel("moments.thumbnail.selectedImage").replace("{name}", selected.name)}
                </Typography>
                <GalleryImagePicker
                  items={gallery}
                  label={getLabel("moments.thumbnail.replacePhoto")}
                  onChoose={(item) =>
                    updateLayer(selected.id, { src: item.url, name: item.filename })
                  }
                />
                <TextField
                  select
                  label={getLabel("moments.thumbnail.frameRatio")}
                  size="small"
                  value={selectedRatioPreset}
                  onChange={(event) => {
                    const value = event.target.value;
                    if (value === "original") {
                      const image = images.get(selected.id);
                      if (image) setImageRatio(image.originalWidth / image.originalHeight);
                    } else if (value !== "custom")
                      setImageRatio(aspectValue(value as ThumbnailDocument["aspect"]));
                  }}
                >
                  <MenuItem value="custom" disabled>
                    {getLabel("moments.thumbnail.custom")}
                  </MenuItem>
                  <MenuItem value="original">
                    {getLabel("moments.thumbnail.originalPhoto")}
                  </MenuItem>
                  {ratioOptions.map((ratio) => (
                    <MenuItem key={ratio} value={ratio}>
                      {ratio}
                    </MenuItem>
                  ))}
                </TextField>
                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: "repeat(auto-fit, minmax(100px, 1fr))",
                    gap: 1,
                  }}
                >
                  {numericField(getLabel("moments.thumbnail.xPercent"), "x")}
                  {numericField(getLabel("moments.thumbnail.yPercent"), "y")}
                  {numericField(getLabel("moments.thumbnail.widthPercent"), "width")}
                  {numericField(getLabel("moments.thumbnail.heightPercent"), "height")}
                  {numericField(getLabel("moments.thumbnail.angle"), "angle")}
                  {numericField(getLabel("moments.thumbnail.zoom"), "zoom", 0.1)}
                  {numericField(getLabel("moments.thumbnail.focusX"), "focusX")}
                  {numericField(getLabel("moments.thumbnail.focusY"), "focusY")}
                </Box>
                <Stack direction={{ xs: "column", sm: "row" }} spacing={1}>
                  <TextField
                    select
                    fullWidth
                    label={getLabel("moments.thumbnail.frame")}
                    size="small"
                    value={selected.frame}
                    onChange={(event) =>
                      updateLayer(selected.id, {
                        frame: event.target.value as ThumbnailLayer["frame"],
                      })
                    }
                  >
                    <MenuItem value="none">{getLabel("moments.thumbnail.noFrame")}</MenuItem>
                    <MenuItem value="polaroid">{getLabel("moments.thumbnail.polaroid")}</MenuItem>
                    <MenuItem value="rounded">{getLabel("moments.thumbnail.rounded")}</MenuItem>
                  </TextField>
                  <TextField
                    select
                    fullWidth
                    label={getLabel("moments.thumbnail.effect")}
                    size="small"
                    value={selected.effect}
                    onChange={(event) =>
                      updateLayer(selected.id, {
                        effect: event.target.value as ThumbnailLayer["effect"],
                      })
                    }
                  >
                    <MenuItem value="none">{getLabel("moments.thumbnail.originalEffect")}</MenuItem>
                    <MenuItem value="grayscale">{getLabel("moments.thumbnail.grayscale")}</MenuItem>
                    <MenuItem value="sepia">{getLabel("moments.thumbnail.sepia")}</MenuItem>
                    <MenuItem value="warm">{getLabel("moments.thumbnail.warm")}</MenuItem>
                    <MenuItem value="grain">{getLabel("moments.thumbnail.grain")}</MenuItem>
                  </TextField>
                </Stack>
                <FormControlLabel
                  control={
                    <Switch
                      checked={selected.shadow}
                      onChange={(_, checked) => updateLayer(selected.id, { shadow: checked })}
                    />
                  }
                  label={getLabel("moments.thumbnail.shadow")}
                />
                <Stack direction="row" spacing={1} flexWrap="wrap">
                  <Button
                    onClick={() =>
                      setDocument((current) => ({
                        ...current,
                        layers: current.layers.filter((layer) => layer.id !== selected.id),
                      }))
                    }
                  >
                    {getLabel("moments.thumbnail.remove")}
                  </Button>
                  <Button
                    onClick={() =>
                      setDocument((current) => {
                        const layers = [...current.layers];
                        const index = layers.findIndex((layer) => layer.id === selected.id);
                        if (index > 0)
                          [layers[index - 1], layers[index]] = [layers[index], layers[index - 1]];
                        return { ...current, layers };
                      })
                    }
                  >
                    {getLabel("moments.thumbnail.sendBack")}
                  </Button>
                  <Button
                    onClick={() =>
                      setDocument((current) => {
                        const layers = [...current.layers];
                        const index = layers.findIndex((layer) => layer.id === selected.id);
                        if (index >= 0 && index < layers.length - 1)
                          [layers[index], layers[index + 1]] = [layers[index + 1], layers[index]];
                        return { ...current, layers };
                      })
                    }
                  >
                    {getLabel("moments.thumbnail.bringForward")}
                  </Button>
                </Stack>
              </Stack>
            </Paper>
          )}
          <Paper sx={{ p: 2 }}>
            <Stack spacing={2}>
              <Typography variant="h6">{getLabel("moments.thumbnail.exportImport")}</Typography>
              <Stack direction="row" spacing={1} flexWrap="wrap">
                <Button
                  variant="contained"
                  disabled={saving || documentValue.layers.length === 0}
                  onClick={() => void saveImage()}
                >
                  {saving
                    ? getLabel("moments.thumbnail.saving")
                    : getLabel("moments.thumbnail.saveCover")}
                </Button>
                <Button startIcon={<DownloadIcon />} onClick={() => void exportImage("png")}>
                  {getLabel("moments.thumbnail.downloadPng")}
                </Button>
                <Button startIcon={<DownloadIcon />} onClick={() => void exportImage("jpeg")}>
                  {getLabel("moments.thumbnail.downloadJpg")}
                </Button>
                <Button onClick={exportJson}>{getLabel("moments.thumbnail.downloadJson")}</Button>
              </Stack>
              <TextField
                multiline
                minRows={3}
                label={getLabel("moments.thumbnail.pasteJson")}
                value={jsonText}
                onChange={(event) => setJsonText(event.target.value)}
              />
              <Button disabled={!jsonText.trim()} onClick={importJson}>
                {getLabel("moments.thumbnail.loadJson")}
              </Button>
            </Stack>
          </Paper>
        </Stack>
      </Stack>
    </Stack>
  );
};
