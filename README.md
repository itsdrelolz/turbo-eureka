
A simulation of an asynchronous messaging queue architecture designed to process file uploads. This project demonstrates the core lifecycle of a messaging queue using Go, PostgreSQL, Redis, and MinIO (S3).

*Note: This is a learning project intended to replicate the mechanics of a job queue. It is strictly limited to PDF files. PDF parsing is often inconsisten and may produce inaccurate results. As a learning project, it intentionally excludes complex production features such as dead letter queues, job loss prevention, and exponential backoff.  

## Prerequisites

* Docker 

1.  **Clone the repository:**
    ```bash
    git clone <REPO_URL>
    cd turbo-eureka
    ```

2.  **Create the Environment File:**
    Create a `.env` file in the root directory based on the example provided.
    ```bash
    cp .env.example .env
    ```

3.  **Start the Services:**
    Run the following command to build and start the application containers:
    ```bash
    docker compose up --build
    ```
    *Wait for the logs to show that the database, MinIO, and API services are ready.*


## Usage

Once the services are running, the API will be available at `http://localhost:8080`.

### 1. Upload a Resume
Upload a PDF file to the processing queue. Replace the file path with the actual location of your PDF.

**Only** `.pdf` files are supported.

```bash
curl -F "resume=@/path/to/your/Resume.pdf" http://localhost:8080/upload
```
Response: You will receive a Job ID in the response (e.g., "019b1f92-dbae-7af4-a229-8499baad07ee").
```bash
curl http://localhost:8080/resumes/<JOB_ID>
```

The result should be the text of the extracted pdf file

## Features to add: 
- Preventing job loss in the event of a crash 
- Job failure retries
