FROM python:3.14-slim AS builder

ENV UV_COMPILE_BYTECODE=1
ENV UV_LINK_MODE=copy
ENV UV_NO_DEV=1
ENV UV_PYTHON_DOWNLOADS=0

RUN apt-get update && apt-get install -y build-essential && rm -rf /var/lib/apt/lists/*
RUN pip install uv

WORKDIR /app
RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=pyproject.toml \
    uv sync --locked --no-install-project
COPY . /app
RUN --mount=type=cache,target=/root/.cache/uv uv sync --locked


FROM python:3.14-slim

RUN groupadd -g 999 nonroot && useradd -m -u 999 -g nonroot nonroot
RUN mkdir -p /home/nonroot/.cache/huggingface/ && chown -R nonroot:nonroot /home/nonroot/.cache/huggingface/
COPY --from=builder --chown=nonroot:nonroot /app /app

ENV PATH="/app/.venv/bin:$PATH"
USER nonroot

WORKDIR /app
CMD ["python", "src/main.py"]
