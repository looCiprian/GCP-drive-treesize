package web

var indexPageTemplate string = `<!doctype html>
<html lang="en">
<title>Drive Space Usage</title>

<head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <!-- Bootstrap CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet"
        integrity="sha384-1BmE4kWBq78iYhFldvKuhfTAU6auU8tT94WrHftjDbrCEXSU1oBoqyl2QvZ6jIW3" crossorigin="anonymous">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/3.7.1/chart.min.js"
        integrity="sha512-QSkVNOCYLtj73J4hbmVoOV6KVZuMluZlioC+trLpewV8qMjsWqlIQvkn1KGX2StWvPMdWGBqim1xlC8krl1EKQ=="
        crossorigin="anonymous" referrerpolicy="no-referrer"></script>
</head>

<body>

    <!-- Optional JavaScript; choose one of the two! -->

    <!-- Option 1: Bootstrap Bundle with Popper -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"
        integrity="sha384-ka7Sk0Gln4gmtz2MlQnikT1wXgYsOg+OMhuP+IlRH9sENBO0LRn5q+8nbTov4+1p"
        crossorigin="anonymous"></script>


    <div class="container">
        <div class="row justify-content-md-center">
            <div class="col-md-auto">
                <a class="btn btn-primary" href="https://one.google.com/storage" target="_blank" role="button">View
                    Storage Stats by Google</a>
                <a class="btn btn-primary" href="https://drive.google.com/" target="_blank" role="button">Open
                    Google
                    Drive</a>
                    <a class="btn btn-primary" href="/stats" role="button">Open
                        View Statistics</a>
            </div>
        </div>
    </div>

    <div class="container">
        <div class="row justify-content-md-center mr-20">
            <div class="col-md-5" style="word-break: break-word;">

                Path:
                {{if .CurrentPath}}
                {{range .CurrentPath}}
                <a href="/node/{{.Id}}">{{.Name}}</a>/
                {{end}}
                {{end}}<br><br>

                {{range .Child}}
                {{if .IsDir}}
                - <a href="/node/{{.Id}}" id="current-node-child">&#128193; Folder: {{.Name}} - {{.HumanSize}}</a><br>
                {{else}}
                - <a href="/node/{{.Id}}" id="current-node-child">&#x1f5ce; File: {{.Name}} - {{.HumanSize}}</a><br>
                {{end}}
                {{end}}

                Go Up:
                <a href="/node/{{.CurrentNode.Parent}}" id="current-node-child">Go Back</a>

            </div>
            <div class="col-md-4" style="word-break: break-word;">
                Info:
                <br>
                <ul>
                    <li>Name: {{.CurrentNode.Name}}</li>
                    <li>Size: {{.CurrentNode.HumanSize}}</li>
                    <li>Open on Google Drive: <a href="{{.CurrentNode.Link}}">{{.CurrentNode.Link}}</a></li>
                    <li>File inside: {{.CurrentNode.NumberOfFiles}}</li>
                    <li>Is Folder: {{.CurrentNode.IsDir}}</li>
                    <li>MimeType: {{.CurrentNode.MimeType}}</li>
                </ul>
                <br>
            </div>
        </div>
    </div>
</body>

</html>`
