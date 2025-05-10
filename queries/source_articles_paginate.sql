SELECT uri,
       json_quote(title),
       pubdate
FROM Articles
WHERE sourceid = :SourceID
ORDER BY pubdate DESC
limit :PageOffset, :PageSize;

