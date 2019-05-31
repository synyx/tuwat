<html>
<head>
<title>Nagios Dashboard</title>
<meta http-equiv="refresh" content="10; URL=foo.php">
<script src="//ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js"></script>
<script src="//netdna.bootstrapcdn.com/twitter-bootstrap/2.2.1/js/bootstrap.min.js"></script>
<link href="//netdna.bootstrapcdn.com/twitter-bootstrap/2.2.1/css/bootstrap-combined.min.css" rel="stylesheet">
<link rel="stylesheet" href="css/blinkftw.css">
<link rel="stylesheet" href="css/main.css">
<script>
  function showInfo(show_data) {
      $("#info-window").fadeIn("fast");
      $("#info-window-text").empty().append(show_data);
  }
  $(document).ready(function() {
      $("#nagioscontainer").load("nagdash.php", function() { $("#spinner").fadeOut("fast"); });
      var refreshId = setInterval(function() {
          <?if ($show_refresh_spinner) {?>
          $("#spinner").fadeIn("fast");
          <? }?>
          $("#nagioscontainer").load("nagdash.php", function() { $("#spinner").fadeOut("fast"); });
      }, 20000);
      $.ajaxSetup({ cache: false });
  });
  $(document).keypress("s", function(e) {
      $("#settings_modal").modal();
  });
</script>
<style type="text/css">
  <?php foreach ($nagios_hosts as $host) { 
        echo ".tag_{$host['tag']}   { background-color: {$host['tagcolour']} }\n"; 
  } ?>
</style>
</head>
<body>
  <div id="spinner"><h3><img src="images/ajax-loader.gif" align="absmiddle"> Refreshing...</h3></div>
  <div id="nagioscontainer"></div>
	<?=build_settings_dialog($nagios_hosts, $unwanted_hosts) ?>
<a href="?known=on">known on</a> - 
<a href="?known=off">known off</a>
</body>
</html>
