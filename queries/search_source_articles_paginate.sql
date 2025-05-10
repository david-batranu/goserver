SELECT uri,
       title,
       pubdate
FROM Articles
WHERE sourceid = :SourceID
  AND title LIKE '%:SearchString%'
  AND id NOT IN
    (SELECT id
     FROM Articles
     WHERE sourceid = :SourceID
     ORDER BY pubdate DESC
     LIMIT :PageOffset)
ORDER BY pubdate DESC
limit :PageSize;

