param(
  [string]$ApiBase  = "http://44.198.15.64:9090/api/v1",
  # Usuario base para login de carga (puede ser el seed carlos@anb.com/password)
  [string]$Email    = "carlos@anb.com",
  [string]$Password = "password",
  [switch]$Insecure,
  [switch]$RunEsc2
)

$ErrorActionPreference = 'Stop'
$HERE = $PSScriptRoot
$SRC  = Join-Path $HERE 'loadtest.go'

if (-not (Test-Path $SRC)) { throw "No encuentro loadtest.go en $HERE." }
$GO = Get-Command go -ErrorAction SilentlyContinue
if (-not $GO) { throw "Go no está en PATH. Instálalo: https://go.dev/dl/" }

# === Helpers ===
function Invoke-GoRun {
  param([string[]]$argv,[string]$label)
  $extra = @()
  if ($Insecure) { $extra += "-insecure" }
  $cmd = @($argv) + $extra
  Write-Host ">> [$label] go run $SRC $($cmd -join ' ')" -ForegroundColor Cyan
  & $GO run $SRC $cmd
  $code = $LASTEXITCODE
  if ($code -eq -1073741510) { Write-Warning "[$label] cancelado por usuario (Ctrl+C/Stop)"; return $false }
  if ($code -ne 0) { throw "[$label] falló (exit $code)" }
  return $true
}
function Save-Json($path, $obj) { $obj | ConvertTo-Json -Depth 10 | Out-File -Encoding utf8 $path }
function First-NonNull($obj, [string[]]$keys) {
  foreach ($k in $keys) { if ($null -ne $obj.$k -and "$($obj.$k)".Length -gt 0) { return $obj.$k } }
  return $null
}
function Get-AnyVideoId {
  param([string]$ApiBase,[string]$token)
  try {
    $p = Invoke-RestMethod -Uri "$ApiBase/videos" -Headers @{ Authorization = "Bearer $token" }
    if ($p -is [System.Array] -and $p.Length -gt 0) { return First-NonNull $p[0] @('id','video_id') }
  } catch {}
  try {
    $q = Invoke-RestMethod -Uri "$ApiBase/public/videos"
    if ($q -is [System.Array] -and $q.Length -gt 0) { return First-NonNull $q[0] @('id','video_id') }
  } catch {}
  return $null
}

# === Carpeta resultados ===
$ts = Get-Date -Format "yyyyMMdd-HHmmss"
$outDir = Join-Path $HERE "resultados-$ts"
New-Item -ItemType Directory -Force -Path $outDir | Out-Null
Write-Host "Resultados en: $outDir"

# === Archivos auxiliares ===
$headersAuth = Join-Path $outDir "headers_auth.txt"
$headersJSON = Join-Path $outDir "headers_json.txt"
"Content-Type: application/json" | Out-File -Encoding ascii $headersJSON

# === AUTH: signup (una vez) ===
# Creamos un email único para no chocar con duplicados
$signupEmail = "loadtest.$ts@anb.com"
$signupBody  = @{ email = $signupEmail; password = "Password123!" } | ConvertTo-Json
try {
  $signupResp = Invoke-RestMethod -Uri "$ApiBase/auth/signup" -Method Post -Body $signupBody -ContentType "application/json"
  Save-Json (Join-Path $outDir "auth_signup_resp.json") $signupResp
  Write-Host "Signup OK → $signupEmail"
} catch {
  Write-Warning "Signup falló (posible duplicado si se re-ejecuta). Continuamos. $_"
}

# === AUTH: login (para token) ===
$loginBody = @{ email=$Email; password=$Password } | ConvertTo-Json
try {
  $loginResp = Invoke-RestMethod -Uri "$ApiBase/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
  Save-Json (Join-Path $outDir "auth_login_resp.json") $loginResp
  $token = $loginResp.access_token
  if (-not $token) { throw "No se obtuvo access_token" }
  "Authorization: Bearer $token" | Out-File -Encoding ascii $headersAuth
  Write-Host "Login de carga OK → $Email"
} catch {
  throw "Login falló. $_"
}

# === LOGIN: carga ligera al endpoint /auth/login (para cubrirlo) ===
$loginBodyPath = Join-Path $outDir "body_login.json"
$loginBody | Out-File -Encoding utf8 $loginBodyPath
$argvLogin = @(
  "-url", "$ApiBase/auth/login",
  "-method", "POST",
  "-headers", $headersJSON,
  "-body", $loginBodyPath,
  "-concurrency", "10",
  "-rate", "10",
  "-duration", "30s",
  "-out_json", (Join-Path $outDir "login_load.json"),
  "-out_csv",  (Join-Path $outDir "login_load.csv")
)
Invoke-GoRun -argv $argvLogin -label "AUTH login (carga ligera)"

# === PUBLIC: GET /public/videos (escenario moderado) ===
$argvPub1 = @(
  "-url", "$ApiBase/public/videos",
  "-method", "GET",
  "-concurrency", "20",
  "-rate", "30",
  "-duration", "2m",
  "-out_json", (Join-Path $outDir "public_videos_esc1.json"),
  "-out_csv",  (Join-Path $outDir "public_videos_esc1.csv")
)
Invoke-GoRun -argv $argvPub1 -label "PUBLIC /public/videos esc1"

# === PRIVATE: GET /videos (escenario moderado) ===
$argvPriv1 = @(
  "-url", "$ApiBase/videos",
  "-method", "GET",
  "-headers", $headersAuth,
  "-concurrency", "20",
  "-rate", "30",
  "-duration", "2m",
  "-out_json", (Join-Path $outDir "videos_esc1.json"),
  "-out_csv",  (Join-Path $outDir "videos_esc1.csv")
)
Invoke-GoRun -argv $argvPriv1 -label "PRIVATE /videos esc1"

# === ESCENARIO 2 (estrés) — SOLO si se pide con -RunEsc2 ===
if ($RunEsc2) {
  # Público: 120 rps, 3 min, conc 60
  $argvEsc2Pub = @(
    "-url", "$ApiBase/public/videos",
    "-method", "GET",
    "-concurrency", "60",
    "-rate", "120",
    "-duration", "3m",
    "-out_json", (Join-Path $outDir "public_videos_esc2.json"),
    "-out_csv",  (Join-Path $outDir "public_videos_esc2.csv")
  )
  Invoke-GoRun -argv $argvEsc2Pub -label "PUBLIC /public/videos esc2"

  # Privado: 100 rps, 3 min, conc 60
  $argvEsc2Priv = @(
    "-url", "$ApiBase/videos",
    "-method", "GET",
    "-headers", $headersAuth,
    "-concurrency", "60",
    "-rate", "100",
    "-duration", "3m",
    "-out_json", (Join-Path $outDir "videos_esc2.json"),
    "-out_csv",  (Join-Path $outDir "videos_esc2.csv")
  )
  Invoke-GoRun -argv $argvEsc2Priv -label "PRIVATE /videos esc2"
}


# === PUBLIC EXTRA: GET /public/rankings (ligero) ===
$argvRank = @(
  "-url", "$ApiBase/public/rankings",
  "-method", "GET",
  "-concurrency", "10",
  "-rate", "20",
  "-duration", "1m",
  "-out_json", (Join-Path $outDir "public_rankings.json"),
  "-out_csv",  (Join-Path $outDir "public_rankings.csv")
)
Invoke-GoRun -argv $argvRank -label "PUBLIC /public/rankings"

# === UPLOAD (una vez) para obtener un video propio y usarlo en el resto de endpoints ===
$muestras = Join-Path $HERE "muestras"
New-Item -ItemType Directory -Force -Path $muestras | Out-Null
$sampleFile = Join-Path $muestras "video_local.mp4"
"video fake" | Out-File -Encoding ascii $sampleFile

# curl.exe para multipart: -F "video=@..."; -F "title=..."
$uploadJson = Join-Path $outDir "upload_resp.json"
$curlAuth = "Authorization: Bearer $token"
$curlCmd  = "curl.exe -s -X POST `"$ApiBase/videos/upload`" -H `"$curlAuth`" -F `"video=@$sampleFile`" -F `"title=LoadTest-$ts`""
Write-Host ">> [UPLOAD] $curlCmd" -ForegroundColor Yellow
$raw = cmd /c $curlCmd
$raw | Out-File -Encoding utf8 $uploadJson
try { $uploadObj = Get-Content $uploadJson | ConvertFrom-Json } catch {}
$videoId = if ($uploadObj) { First-NonNull $uploadObj @('id','video_id') } else { $null }

# Si no logramos video_id del upload, tratamos de conseguir cualquiera
if (-not $videoId) {
  $videoId = Get-AnyVideoId -ApiBase $ApiBase -token $token
}
if (-not $videoId) { Write-Warning "No se obtuvo video_id; se saltarán endpoints por id." }

# === PRIVATE BY ID: GET /videos/:id (ligero)
if ($videoId) {
  $argvGetById = @(
    "-url", "$ApiBase/videos/$videoId",
    "-method", "GET",
    "-headers", $headersAuth,
    "-concurrency", "10",
    "-rate", "20",
    "-duration", "1m",
    "-out_json", (Join-Path $outDir "video_${videoId}_get.json"),
    "-out_csv",  (Join-Path $outDir "video_${videoId}_get.csv")
  )
  Invoke-GoRun -argv $argvGetById -label "PRIVATE /videos/:id"
}

# === PRIVATE DOWNLOAD: GET /videos/:id/download (ligero)
if ($videoId) {
  $argvDownload = @(
    "-url", "$ApiBase/videos/$videoId/download",
    "-method", "GET",
    "-headers", $headersAuth,
    "-concurrency", "10",
    "-rate", "20",
    "-duration", "1m",
    "-out_json", (Join-Path $outDir "video_${videoId}_download.json"),
    "-out_csv",  (Join-Path $outDir "video_${videoId}_download.csv")
  )
  Invoke-GoRun -argv $argvDownload -label "PRIVATE /videos/:id/download"
}

# === PRIVATE MARK-PROCESSED: POST /videos/:id/mark-processed (ligero)
if ($videoId) {
  $argvMark = @(
    "-url", "$ApiBase/videos/$videoId/mark-processed",
    "-method", "POST",
    "-headers", $headersAuth,
    "-concurrency", "5",
    "-rate", "5",
    "-duration", "30s",
    "-out_json", (Join-Path $outDir "video_${videoId}_mark.json"),
    "-out_csv",  (Join-Path $outDir "video_${videoId}_mark.csv")
  )
  Invoke-GoRun -argv $argvMark -label "PRIVATE /videos/:id/mark-processed"
}

# === VOTES: POST /public/videos/:id/vote (ligero) + DELETE vote (ligero)
if ($videoId) {
  $argvVote = @(
    "-url", "$ApiBase/public/videos/$videoId/vote",
    "-method", "POST",
    "-headers", $headersAuth,
    "-concurrency", "5",
    "-rate", "5",
    "-duration", "30s",
    "-out_json", (Join-Path $outDir "vote_${videoId}_post.json"),
    "-out_csv",  (Join-Path $outDir "vote_${videoId}_post.csv")
  )
  Invoke-GoRun -argv $argvVote -label "VOTE POST /public/videos/:id/vote"

  $argvUnvote = @(
    "-url", "$ApiBase/public/videos/$videoId/vote",
    "-method", "DELETE",
    "-headers", $headersAuth,
    "-concurrency", "5",
    "-rate", "5",
    "-duration", "30s",
    "-out_json", (Join-Path $outDir "vote_${videoId}_delete.json"),
    "-out_csv",  (Join-Path $outDir "vote_${videoId}_delete.csv")
  )
  Invoke-GoRun -argv $argvUnvote -label "VOTE DELETE /public/videos/:id/vote"
}

# === PRIVATE DELETE /videos/:id (una vez, al final)
if ($videoId) {
  try {
    $delResp = Invoke-RestMethod -Uri "$ApiBase/videos/$videoId" -Method Delete -Headers @{ Authorization = "Bearer $token" }
    Save-Json (Join-Path $outDir "video_${videoId}_delete_resp.json") $delResp
    Write-Host "DELETE /videos/$videoId OK (1 vez)" -ForegroundColor Green
  } catch {
    Write-Warning "DELETE /videos/$videoId falló (revisa permisos/estado). $_"
  }
}

Write-Host "`n✅ Prueba completa (todos los endpoints) finalizada. Carpeta: $outDir"
