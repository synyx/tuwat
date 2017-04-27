<html><head>
<meta http-equiv="refresh" content="15; URL=http://contargo-nagdash.synyx.coffee/">
</head>
<link rel="stylesheet" href="css/blinkftw.css">
<link rel="stylesheet" href="css/main.css">
<body>

<?php
$of = "done.json";
$op = "progress.json";
$open = fopen($of, "r");
$progress = fopen($op, "r");
$d = json_decode(fread($open, filesize($of)), JSON_PRETTY_PRINT);
$p = json_decode(fread($progress, filesize($op)), JSON_PRETTY_PRINT);
fclose($open);
fclose($progress);


#echo(var_dump($d));
$data = "<table align=\"left\"><tr><td colspan=2><b>Done!</b></td></tr>\r\n";

foreach($d as &$gu){

#echo ">>>>".var_dump($gu)."<<<<";
	#var_dump($gu);
	foreach ($gu as $lala){
#var_dump($lala);

$hero = $lala['assigned_to']['name'];
if (strlen($hero) == 0){
	$hero = "Shy guy";
}

$data .= "<tr><td>".$lala['subject']."</td><td>Hero: <b>". $hero."</b></td></tr>\r\n";

#var_dump($lala['assigned_to']['name']);
#

#die($data);
}
}

#echo $data;
$progress = 0;
#var_dump($p);
foreach($p as &$gu){


#echo ">>>>".var_dump($gu)."<<<<";
	#var_dump($gu);
	foreach ($gu as $lala){
#var_dump($lala);

$hero = $lala['assigned_to']['name'];
if (strlen($hero) == 0){
	$hero = "Shy guy";
}
$progress++;

$pdata .= "<tr><td>".$lala['subject']."</td><td>Worker-Node: <b>". $hero."</b></td></tr>\r\n";
#$data .= $lala['subject']."\t\t In Progress by: ". $hero."\r\n";

#var_dump($lala['assigned_to']['name']);
#

#die($data);
}
}
$data .= "<table><tr><td colspan=2><b>In Progress ($progress)~</b></td></tr>\r\n";
$data .= $pdata;
echo $data;
echo "</table>";
?>
