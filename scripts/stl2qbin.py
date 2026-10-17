#!/usr/bin/env python3
"""Convert a binary or ASCII STL mesh into an indexed 16-bit quantized binary (.bin).

Binary layout (little-endian, 32-byte header):
  Offset  Size   Type         Description
  0       4      char[4]      Magic bytes: "GQ16"
  4       2      uint16       Unique vertex count (V)
  6       2      uint16       Triangle count (T)
  8       12     float32[3]   Bounding-box minimum (minX, minY, minZ)
  20      12     float32[3]   Bounding-box size (sizeX, sizeY, sizeZ)
  32      V*6    uint16[V*3]  Quantized vertex coordinates (qx, qy, qz) in 0..65535
  32+V*6  T*6    uint16[T*3]  Triangle vertex indices (i0, i1, i2)
"""

from __future__ import annotations

import argparse
from collections import defaultdict
from pathlib import Path
import struct
import sys

MAGIC = b"GQ16"
DEFAULT_INPUT = Path("asset/go_gopher_low_0.02.stl")
DEFAULT_OUTPUT = Path("asset/go_gopher_low_0.02.bin")


def parse_stl(data: bytes) -> list[tuple[tuple[float, float, float], ...]]:
    """Parse binary or ASCII STL bytes into a list of triangles (each 3 (x,y,z) tuples)."""
    if len(data) >= 84:
        tri_count = struct.unpack_from("<I", data, 80)[0]
        expected_size = 84 + tri_count * 50
        if len(data) == expected_size:
            triangles = []
            for i in range(tri_count):
                off = 84 + i * 50
                # Skip normal (3 floats), unpack 3 vertices (9 floats)
                v = struct.unpack_from("<9f", data, off + 12)
                triangles.append((v[0:3], v[3:6], v[6:9]))
            return triangles

    # Fallback: ASCII STL
    text = data.decode("utf-8", errors="replace")
    verts: list[tuple[float, float, float]] = []
    for line in text.splitlines():
        parts = line.strip().split()
        if len(parts) == 4 and parts[0] == "vertex":
            verts.append((float(parts[1]), float(parts[2]), float(parts[3])))
    if not verts or len(verts) % 3 != 0:
        raise ValueError("Unsupported or corrupted STL file")
    return [(verts[i], verts[i + 1], verts[i + 2]) for i in range(0, len(verts), 3)]


def convert_stl_to_qbin(src: Path, dst: Path, precision: int = 5) -> None:
    raw = src.read_bytes()
    triangles = parse_stl(raw)
    if not triangles:
        raise ValueError(f"No triangles found in {src}")

    unique_verts: list[tuple[float, float, float]] = []
    vert_index: dict[tuple[float, float, float], int] = {}
    indices: list[int] = []

    for tri in triangles:
        for x, y, z in tri:
            key = (round(x, precision), round(y, precision), round(z, precision))
            idx = vert_index.get(key)
            if idx is None:
                idx = len(unique_verts)
                vert_index[key] = idx
                unique_verts.append((x, y, z))
            indices.append(idx)

    # Ensure each connected component has outward-facing winding (positive signed volume).
    adj: dict[int, set[int]] = defaultdict(set)
    for i in range(0, len(indices), 3):
        a, b, c = indices[i], indices[i + 1], indices[i + 2]
        adj[a].update((b, c))
        adj[b].update((a, c))
        adj[c].update((a, b))

    visited: set[int] = set()
    for start_v in range(len(unique_verts)):
        if start_v in visited or start_v not in adj:
            continue
        comp_verts: set[int] = set()
        stack = [start_v]
        visited.add(start_v)
        while stack:
            curr = stack.pop()
            comp_verts.add(curr)
            for nb in adj[curr]:
                if nb not in visited:
                    visited.add(nb)
                    stack.append(nb)

        signed_vol6 = 0.0
        comp_tris: list[int] = []
        for ti in range(0, len(indices), 3):
            if indices[ti] in comp_verts:
                comp_tris.append(ti)
                x0, y0, z0 = unique_verts[indices[ti]]
                x1, y1, z1 = unique_verts[indices[ti + 1]]
                x2, y2, z2 = unique_verts[indices[ti + 2]]
                signed_vol6 += (
                    x0 * (y1 * z2 - y2 * z1)
                    - x1 * (y0 * z2 - y2 * z0)
                    + x2 * (y0 * z1 - y1 * z0)
                )
        if signed_vol6 < -1e-2:
            for ti in comp_tris:
                indices[ti + 1], indices[ti + 2] = indices[ti + 2], indices[ti + 1]

    v_count = len(unique_verts)
    t_count = len(triangles)
    if v_count > 65535 or t_count > 65535:
        raise ValueError(
            f"Mesh exceeds uint16 limits: {v_count} vertices, {t_count} triangles"
        )

    xs = [v[0] for v in unique_verts]
    ys = [v[1] for v in unique_verts]
    zs = [v[2] for v in unique_verts]

    min_x, max_x = min(xs), max(xs)
    min_y, max_y = min(ys), max(ys)
    min_z, max_z = min(zs), max(zs)

    size_x = (max_x - min_x) or 1.0
    size_y = (max_y - min_y) or 1.0
    size_z = (max_z - min_z) or 1.0

    q_coords: list[int] = []
    for x, y, z in unique_verts:
        qx = max(0, min(65535, int(round((x - min_x) / size_x * 65535))))
        qy = max(0, min(65535, int(round((y - min_y) / size_y * 65535))))
        qz = max(0, min(65535, int(round((z - min_z) / size_z * 65535))))
        q_coords.extend((qx, qy, qz))

    header = struct.pack(
        "<4sHH6f",
        MAGIC,
        v_count,
        t_count,
        min_x,
        min_y,
        min_z,
        size_x,
        size_y,
        size_z,
    )
    vert_buf = struct.pack(f"<{len(q_coords)}H", *q_coords)
    idx_buf = struct.pack(f"<{len(indices)}H", *indices)

    out_bytes = header + vert_buf + idx_buf
    dst.parent.mkdir(parents=True, exist_ok=True)
    dst.write_bytes(out_bytes)

    ratio = (1.0 - len(out_bytes) / len(raw)) * 100.0
    print(
        f"{src} ({len(raw):,} B) -> {dst} ({len(out_bytes):,} B, -{ratio:.1f}%) "
        f"[vertices: {t_count * 3:,} -> {v_count:,}, triangles: {t_count:,}]"
    )


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Convert STL mesh to indexed 16-bit quantized binary (.bin)."
    )
    parser.add_argument(
        "input",
        nargs="?",
        type=Path,
        default=DEFAULT_INPUT,
        help=f"Input STL path (default: {DEFAULT_INPUT})",
    )
    parser.add_argument(
        "output",
        nargs="?",
        type=Path,
        default=DEFAULT_OUTPUT,
        help=f"Output quantized binary path (default: {DEFAULT_OUTPUT})",
    )
    args = parser.parse_args()

    try:
        convert_stl_to_qbin(args.input, args.output)
    except Exception as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
