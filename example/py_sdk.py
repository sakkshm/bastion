from bastion import Bastion
import threading
import time

def test_sessions(sdk: Bastion):
    print("\n=== SESSION TESTS ===")

    # create session
    session = sdk.sessions.create()
    session_id = session["session_id"]
    print("Created:", session)

    # status before start
    status = sdk.sessions.status(session_id)
    print("Initial status:", status)

    # start session
    started = sdk.sessions.start(session_id)
    print("Started:", started)

    # status after start
    status = sdk.sessions.status(session_id)
    print("Running status:", status)

    # stop session
    stopped = sdk.sessions.stop(session_id)
    print("Stopped:", stopped)

    # restart session
    restarted = sdk.sessions.start(session_id)
    print("Restarted:", restarted)

    # delete session
    deleted = sdk.sessions.delete(session_id)
    print("Deleted:", deleted)

    # edge case: invalid session
    try:
        sdk.sessions.status("invalid-session-id")
    except Exception as e:
        print("Expected session error:", type(e).__name__, e)


def test_files(sdk: Bastion, session_id: str):
    print("\n=== FILE TESTS ===")

    # upload file
    file_content = b"hello bastion sdk test"
    upload = sdk.files.upload(session_id, file_content, "/hello/test.txt")
    print("Upload:", upload)

    # list files
    listing = sdk.files.list(session_id, "/hello")
    print("List:", listing)

    # delete file
    deleted = sdk.files.delete(session_id, "/hello/test.txt")
    print("Delete:", deleted)

    # edge case: delete missing file
    try:
        sdk.files.delete(session_id, "/does/not/exist.txt")
    except Exception as e:
        print("Expected file error:", type(e).__name__, e)


def test_jobs(sdk: Bastion, session_id: str):
    print("\n=== JOB TESTS ===")

    # run job
    job = sdk.jobs.run(session_id, ["echo", "hello"])
    job_id = job["job_id"]
    print("Job created:", job)

    # get job
    status = sdk.jobs.get(session_id, job_id)
    print("Job status:", status)

    # wait job
    final = sdk.jobs.wait(session_id, job_id, timeout=10)
    print("Job finished:", final)

    # run and wait
    result = sdk.jobs.run_and_wait(session_id, ["echo", "run_and_wait"])
    print("run_and_wait:", result)

    # watch job
    def cb(update):
        print("WATCH UPDATE:", update["status"])

    job2 = sdk.jobs.run(session_id, ["sleep", "1"])
    sdk.jobs.watch(session_id, job2["job_id"], cb)

    # edge case: timeout
    try:
        slow_job = sdk.jobs.run(session_id, ["sleep", "999"])
        sdk.jobs.wait(session_id, slow_job["job_id"], timeout=1)
    except TimeoutError as e:
        print("Expected timeout:", e)

    # edge case: invalid job
    try:
        sdk.jobs.get(session_id, "invalid-job")
    except Exception as e:
        print("Expected job error:", type(e).__name__, e)

def test_terminal(sdk: Bastion, session_id: str):
    print("\n=== TERMINAL TESTS ===")

    terminal = sdk.terminal

    def on_message(msg):
        print("WS MESSAGE:", msg)

    def on_open():
        print("WS OPENED")

    def on_close():
        print("WS CLOSED")

    def on_error(e):
        print("WS ERROR:", e)

    # run connect in background thread (non-blocking)
    t = threading.Thread(
        target=terminal.connect,
        kwargs={
            "session_id": session_id,
            "on_message": on_message,
            "on_open": on_open,
            "on_close": on_close,
            "on_error": on_error,
        },
        daemon=True,
    )
    t.start()

    # wait for connection + init packet
    time.sleep(2)

    # send input
    try:
        terminal.send_input("ls -la\n")
    except Exception as e:
        print("send_input error:", type(e).__name__, e)

    # exec command
    try:
        terminal.exec("echo hello from exec")
    except Exception as e:
        print("exec error:", type(e).__name__, e)

    # let messages stream
    time.sleep(3)

    terminal.close()
    print("Terminal test complete")

def main():
    sdk = Bastion(
        base_url="http://localhost:8080",
        api_key="bastion_I6Sska_FMXEo2N6M90vAjwJWhWKpJG2DTvtKuUd",
    )

    try:
        # sessions lifecycle
        session = sdk.sessions.create()["session_id"]
        sdk.sessions.start(session)

        # run full test suite
        test_sessions(sdk)
        test_files(sdk, session)
        test_jobs(sdk, session)
        test_terminal(sdk, session)

    except Exception as e:
        print("FATAL ERROR:", type(e).__name__, e)

    finally:
        sdk.close()


if __name__ == "__main__":
    main()