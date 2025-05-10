SELECT uri,
       title,
       pubdate
FROM Articles
WHERE title LIKE '%:SearchString%'
ORDER BY pubdate DESC
limit :PageOffset, :PageSize;

