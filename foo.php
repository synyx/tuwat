<html>
<head>
<title>Nagios Dashboard</title>
<meta http-equiv="refresh" content="90">
<script src="//ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js"></script>
<script src="//netdna.bootstrapcdn.com/twitter-bootstrap/2.2.1/js/bootstrap.min.js"></script>
<link href="//netdna.bootstrapcdn.com/twitter-bootstrap/2.2.1/css/bootstrap-combined.min.css" rel="stylesheet">
<link rel="stylesheet" href="css/blinkftw.css">
<link rel="stylesheet" href="css/main.css">

<?php
date_default_timezone_set('Europe/Berlin');
if (date("G") >= 20){
	echo "<marquee><center><blink><h1>Alles wird gut! &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; echt!</h1></blink></center></marquee>";
}
?>

<?php 
flush();
include("nagdash.php");
