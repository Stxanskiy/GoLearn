# Sandbox for the Python path: the interpreter, PyQt6 and psycopg.
#
# Not built on golearn/sandbox: that image exists for the DevOps courses and
# carries an offline apt repository, a systemctl shim and a set of CLI tools a
# Python lesson never touches. This one only needs an interpreter.
#
# Debian bookworm (python:3.12-slim) is glibc 2.36; the PyQt6 wheels are tagged
# manylinux_2_34, so they install. Alpine is not an option — there are no musl
# wheels for PyQt6 and building Qt from source is hours.
FROM python:3.12-slim

# Qt carries its own libraries but links against these. libgl1 and libegl1 are
# what the classic "ImportError: libGL.so.1" is about; libxkbcommon0 is needed
# by QtGui even with no X server; without the fonts, text metrics and anything
# rendered come out wrong.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      libgl1 libegl1 \
      libglib2.0-0 libdbus-1-3 \
      libxkbcommon0 \
      libfontconfig1 libfreetype6 fonts-dejavu-core \
      ca-certificates \
 && rm -rf /var/lib/apt/lists/*

# Versions are pinned: a course is written against what the student runs, and a
# silent minor upgrade is how a lesson's screenshots stop matching reality.
RUN pip install --no-cache-dir \
      PyQt6==6.11.0 \
      pytest==8.3.4 \
      pytest-qt==4.5.0 \
      "psycopg[binary]==3.3.6"

# QT_QPA_PLATFORM is deliberately NOT set here. A text task runs with no display
# at all; a GUI task is started with `QT_QPA_PLATFORM=vnc`, which makes the
# student's own program the VNC server (plugins/platforms/libqvnc.so ships in
# the wheel). Baking a default would hide which mode a lesson is in.
ENV XDG_RUNTIME_DIR=/tmp/runtime-root \
    QT_LOGGING_RULES=qt.qpa.*=false \
    QT_ACCESSIBILITY=0 \
    PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1
RUN mkdir -p /tmp/runtime-root && chmod 700 /tmp/runtime-root

WORKDIR /root
CMD ["sleep", "infinity"]
