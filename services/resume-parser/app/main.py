from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()


class ParseRequest(BaseModel):
    resume_text: str


class ParseResponse(BaseModel):
    skills: list[str]
    summary: str


@app.get("/healthz")
async def healthz() -> dict[str, str]:
    return {"status": "ok", "service": "resume-parser"}


@app.post("/parse", response_model=ParseResponse)
async def parse_resume(payload: ParseRequest) -> ParseResponse:
    words = [word.strip(",.") for word in payload.resume_text.split()]
    skills = sorted({word for word in words if word.istitle()})
    summary = f"Extracted {len(skills)} potential skills."
    return ParseResponse(skills=skills, summary=summary)
