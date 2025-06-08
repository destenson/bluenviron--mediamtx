

# Find all the video files in the recordings direcotory and print out their lengths (with frame counts)
$recordingsPath = "..\recordings"

$videoFiles = Get-ChildItem -Path $recordingsPath -Filter "*.mp4" -Recurse

# define a function to get the length of a video file using ffprobe
function Get-VideoLength {
    param (
        [string]$filePath
    )

    $ffprobePath = "ffprobe"  # Ensure ffprobe is in your PATH or provide the full path
    $output = & $ffprobePath -v error -select_streams v:0 -show_entries stream=duration -of default=noprint_wrappers=1:nokey=1 $filePath
    return [double]$output
}

foreach ($file in $videoFiles) {
    $filePath = $file.FullName
    $mediaInfo = Get-Item $filePath | Select-Object -ExpandProperty Length
    $frameCount = Get-VideoLength -filePath $filePath

    Write-Output "File: $($file.Name), Byte Length: $mediaInfo, Video Length: $frameCount seconds"
}

