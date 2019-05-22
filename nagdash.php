<?php
@session_start();
error_reporting(E_ALL ^ E_NOTICE);
require_once 'config.php';
require_once 'utils.php';
require_once 'timeago.php';

if (!function_exists('curl_init')) {
  die("ERROR: The PHP curl extension must be installed for Nagdash to function");
}

$nagios_host_status = array(0 => "UP", 1 => "DOWN", 2 => "UNREACHABLE");
$nagios_service_status = array(0 => "OK", 1 => "WARNING", 2 => "CRITICAL", 3 => "UNKNOWN");
$nagios_host_status_colour = array(0 => "status_green", 1 => "status_red", 2 => "status_yellow");
$nagios_service_status_colour = array(0 => "status_green", 1 => "status_yellow", 2 => "status_red", 3 => "status_grey");
$nagios_state_type = array(0 => "SOFT", 1 => "HARD");

$nagios_toggle_status = array(0 => "disabled", 1 => "enabled");

$sort_by_time = ( isset($sort_by_time) && $sort_by_time ) ? true : false;

$errors = array();
$state = array();
$host_summary = array();
$service_summary = array();
$down_hosts = array();
$known_hosts = array();
$known_services = array();
$broken_services = array();
$curl_stats = array();

// HAHA,  synyx was here :D
//

$ignore_service[] = "Debian Updates";
#$ignore_service[] = "Puppet Agent Check";
$ignore_service[] = "Redhat Updates";
$ignore_service[] = "updates";
$ignore_service[] = "Passive BACKUP check";
#$ignore_service[] = "Passive Backup";

function ignore($service_name){
  global $ignore_service;
  $service_val = $service_name;
  ## Soll TRUE zurückgeben wenn Service nicht $ignore_service ist.
  foreach ($ignore_service as $replace){
    $service_val = str_replace($replace, "", $service_val);
  }
  if ($service_val == $service_name){
    return TRUE;
  } else {
    return FALSE;
  }
}


// Function that does the dirty to connect to the Nagios API
function connectNagiosApi($hostname, $port, $protocol) {

    global $curl_stats;

    $ch = curl_init("{$protocol}://{$hostname}:{$port}/state");
    curl_setopt($ch, CURLOPT_ENCODING, 'gzip');
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    if (!$json = curl_exec($ch)) {
        return "<pre>Attempt to hit API failed, sorry. Curl said: " . curl_error($ch) . "</pre>";
    } else {
        $curl_stats["$hostname:$port"] = curl_getinfo($ch);
    }
    curl_close($ch);

    if (!$state = json_decode($json, true)) {
        return "Attempt to hit API failed, sorry (JSON decode failed)";
    }
    $curl_stats["$hostname:$port"]['objects'] = count($state['content']);
    return $state['content'];
}

function connectIcinga2($url) {

    global $curl_stats;

    $state = icinga2v1Get($url, 'objects/hosts');
    $hosts = array();
    if (isset($state['results'])) {
      foreach ($state['results'] as $host) {
          $hosts[$host['name']] = $host['attrs'];
          $hosts[$host['name']]['services'] = array();
          $hosts[$host['name']]['downtimes'] = array();
          $hosts[$host['name']]['current_state'] = $host['attrs']['state'];
          $hosts[$host['name']]['last_state_change'] = $host['attrs']['last_state_change'];
          $hosts[$host['name']]['problem_has_been_acknowledged'] = $host['attr']['acknowledgement'];
          $hosts[$host['name']]['scheduled_downtime_depth'] = $host['attrs']['downtime_depth'];
          $hosts[$host['name']]['notifications_enabled'] = $host['attrs']['enable_notifications'] ? 1 : 0;
          $hosts[$host['name']]['plugin_output'] = $host['attrs']['output'];
          $hosts[$host['name']]['current_attempt'] = $host['attrs']['check_attempt'];
          $hosts[$host['name']]['max_attempts'] = $host['attrs']['max_check_attempts'];
          $hosts[$host['name']]['state_type'] = $host['attrs']['state_type'];
      }
    } else {
      echo $state[0];
    }

    $state = icinga2v1Get($url, 'objects/services');
    if (isset($state['results'])) {
      foreach ($state['results'] as $service) {
          $hn = $service['attrs']['host_name'];
          $sn = $service['attrs']['display_name'];
          $hosts[$hn]['services'][$sn] = $service['attrs'];
          $hosts[$hn]['services'][$sn]['downtimes'] = array();
          $hosts[$hn]['services'][$sn]['current_state'] = $service['attrs']['state'];
          $hosts[$hn]['services'][$sn]['problem_has_been_acknowledged'] = $service['attrs']['acknowledgement'];
          $hosts[$hn]['services'][$sn]['scheduled_downtime_depth'] = $service['attrs']['downtime_depth'];
          $hosts[$hn]['services'][$sn]['notifications_enabled'] = $service['attrs']['enable_notifications'] ? 1 : 0;
          $hosts[$hn]['services'][$sn]['plugin_output'] = $service['attrs']['last_check_result']['output'];
          $hosts[$hn]['services'][$sn]['max_attempts'] = $service['attrs']['max_check_attempts'];
          $hosts[$hn]['services'][$sn]['current_attempt'] = $service['attrs']['check_attempt'];
          $hosts[$hn]['services'][$sn]['state_type'] = $service['attrs']['state_type'];
      }
    } else {
      echo $state[0];
    }

    $state = icinga2v1Get($url, 'objects/downtimes');
    if (isset($state['results'])) {
      foreach ($state['results'] as $downtime) {
          $hn = $downtime['attrs']['host_name'];
          $sn = $downtime['attrs']['service_name'];
          if ($sn == "") {
              // host downtime
              $hosts[$hn]['downtimes'][] = $downtime;
          } else {
              // service downtime
              $hosts[$hn]['services'][$sn][] = $downtime;
          }
      }
    } else {
      echo $state[0];
    }

    return $hosts;
}
function icinga2v1Get($url, $endpoint) {
    $hostname = parse_url($url, PHP_URL_HOST);
    $port = parse_url($url, PHP_URL_PORT);
    $username = parse_url($url, PHP_URL_USER);
    $password = parse_url($url, PHP_URL_PASS);
    $request_url = "$url/v1/{$endpoint}";
    $headers = array(
            'Accept: application/json',
            'X-HTTP-Method-Override: GET'
    );
    $ch = curl_init();
    curl_setopt_array($ch, array(
        CURLOPT_URL => $request_url,
        CURLOPT_HTTPHEADER => $headers,
        CURLOPT_USERPWD => $username . ":" . $password,
        CURLOPT_RETURNTRANSFER => true,
        #CURLOPT_CAINFO => "icinga.synyx.coffee.ca.crt", //re-use the icinga2 master ca.crt
        #CURLOPT_SSL_VERIFYHOST => 2,
        #CURLOPT_SSL_VERIFYPEER => 1
        CURLOPT_SSL_VERIFYHOST => 0,
        CURLOPT_SSL_VERIFYPEER => 0
    ));
    if (!$json = curl_exec($ch)) {
        return ["<pre>Attempt to hit API failed, sorry. Curl said: " . curl_error($ch) . "</pre>"];
    } else {
        $curl_stats["$hostname:$port"] = curl_getinfo($ch);
    }
    curl_close($ch);

    if (!$state = json_decode($json, true)) {
        return ["Attempt to hit API failed, sorry (JSON decode failed)"];
    }
    $curl_stats["$hostname:$port"]['objects'] += count($state['results']);

    return $state;
}

// Check to see if the user has a cookie that disables some hosts
$unwanted_hosts = unserialize($_COOKIE['nagdash_unwanted_hosts']);
if (!is_array($unwanted_hosts)) $unwanted_hosts = array();

// Collect the API data from each Nagios host.
foreach ($nagios_hosts as $host) {
    // Check if the host has been disabled locally
    if (!in_array($host['tag'][0], $unwanted_hosts)) {
        if ($host['type'] == 'icinga2') {
            $host_state = connectIcinga2($host['url']);
        } else {
            $host_state = connectNagiosApi($host['hostname'], $host['port'], $host['protocol']);
        }
        if (is_string($host_state)) {
            $errors[] = "Could not connect to API on host {$host['hostname']}, port {$host['port']}: {$host_state}";
        } else {
            foreach ($host_state as $this_host => $null) {

                if (isset($state[$this_host])) {
                  $state[$this_host]['tag'][] = $host['tag'];
                  # Jo: mostly trust what's there.. aechtz.
                  foreach ($null['services'] as $svcname => $svc) {
                    if (!isset($state[$this_host]['services'][$svcname])) {
                      $state[$this_host]['services'][$svcname] = $svc;
                      $state[$this_host]['services'][$svcname]['tag'] = array($host['tag']);
                    } else {
                      $state[$this_host]['services'][$svcname]['tag'][] = $host['tag'];
                    }
                  }
                } else {
                  $state[$this_host] = $null;

                  foreach ($state[$this_host]['services'] as $svcname => $scv) {
                    $state[$this_host]['services'][$svcname]['tag'] = array($host['tag']);
                  }
                  $state[$this_host]['tag'] = array($host['tag']);
                }

            }
        }
    }
}

if (isset($mock_state_file)) {
  $data = json_decode(file_get_contents($mock_state_file), true);
  $state = $data['content'];
}

// Sort the array alphabetically by hostname.
deep_ksort($state);

// At this point, the data collection is completed.

if (count($errors) > 0) {
    foreach ($errors as $error) {
        echo "<div class='status_red'>{$error}</div>";
    }
}
foreach($state as $hostname => $host_detail) {
    // Check if the host matches the filter
    if (preg_match("/$filter/", $hostname)) {
        // If the host is NOT OK...
        if ($host_detail['current_state'] != 0) {
            // Sort the host into the correct array. It's either a known issue or not.
            if ( ($host_detail['problem_has_been_acknowledged'] > 0) || ($host_detail['scheduled_downtime_depth'] > 0) || ($host_detail['notifications_enabled'] == 0) ) {
                $array_name = "known_hosts";
            } else {
                $array_name = "down_hosts";
            }

            // Populate the array.
            array_push($$array_name, array(
                "hostname" => $hostname,
                "host_state" => $host_detail['current_state'],
                "duration" => timeago($host_detail['last_state_change'], null, null, false),
                "detail" => $host_detail['plugin_output'],
                "current_attempt" => $host_detail['current_attempt'],
                "max_attempts" => $host_detail['max_attempts'],
                "tag" => $host_detail['tag'],
                "is_hard" => ($host_detail['current_attempt'] >= $host_detail['max_attempts'] || $host_detail['state_type'] == 1) ? true : false,
                "is_downtime" => ($host_detail['scheduled_downtime_depth'] > 0) ? true : false,
                "is_ack" => ($host_detail['problem_has_been_acknowledged'] > 0) ? true : false,
                "is_enabled" => ($host_detail['notifications_enabled'] > 0) ? true : false,
            ));
        }

        // In any case, increment the overall status counters.
        $host_summary[$host_detail['current_state']]++;

        // Now parse the statuses for this host.
        foreach($host_detail['services'] as $service_name => $service_detail) {
            // If the host is OK, AND the service is NOT OK.
            if ($service_detail['current_state'] != 0 && $host_detail['current_state'] == 0) {
                // Sort the service into the correct array. It's either a known issue or not.
                if ( ($service_detail['problem_has_been_acknowledged'] > 0) || ($service_detail['scheduled_downtime_depth'] > 0) || ( $service_detail['notifications_enabled'] == 0 ) ||
                        ($host_detail['scheduled_downtime_depth'] > 0) ) {
                    $array_name = "known_services";
                } else {
                    $array_name = "broken_services";
                }
                $downtime_remaining = null;
                $downtimes = array_merge($service_detail['downtimes'], $host_detail['downtimes']);
                if ($host_detail['scheduled_downtime_depth'] > 0 || $service_detail['scheduled_downtime_depth'] > 0) {
                    if (count($downtimes) > 0) {
                        $downtime_info = array_pop($downtimes);
                        $downtime_remaining = "- ". timeago($downtime_info['end_time'], null, null, false) . " left";
                    }
                }
                array_push($$array_name, array(
                    "hostname" => $hostname,
                    "service_name" => $service_name,
                    "service_state" => $service_detail['current_state'],
                    "duration" => timeago($service_detail['last_state_change'], null, null, false),
                    "last_state_change" => $service_detail['last_state_change'],
                    "detail" => $service_detail['plugin_output'],
                    "current_attempt" => $service_detail['current_attempt'],
                    "max_attempts" => $service_detail['max_attempts'],
                    "tag" => $service_detail['tag'],
                    "is_hard" => ($service_detail['current_attempt'] >= $service_detail['max_attempts'] || $service_detail['state_type'] == 1) ? true : false,
                    "is_downtime" => ($service_detail['scheduled_downtime_depth'] > 0 || $host_detail['scheduled_downtime_depth'] > 0) ? true : false,
                    "downtime_remaining" => $downtime_remaining,
                    "is_ack" => ($service_detail['problem_has_been_acknowledged'] > 0) ? true : false,
                    "is_enabled" => ($service_detail['notifications_enabled'] > 0) ? true : false,
                ));

            }
            if ($host_detail['current_state'] == 0) {
                $service_summary[$service_detail['current_state']]++;
            }
        }
    }
}
ksort($host_summary);
ksort($service_summary);
if (empty($host_summary)) { ?>
  <div class="frame">
      <div class="section">
        <div class="header">
          <h3>Host status</h3>
        </div>
        <table class="widetable status_red"><tr><td><b>No Hosts given</b></td></tr></table>
  <?php
} else {
  ?>
  <div id="info-window"><button class="close" onClick='$("#info-window").fadeOut("fast");'>&times;</button><div id="info-window-text"></div></div>
  <div class="frame">
      <div class="section">
        <div class="header">
          <h3>Host status</h3>
          <p class="totals"><b>Total:</b> <?php foreach($host_summary as $state => $count) { echo "<span class='{$nagios_host_status_colour[$state]}'>{$count}</span> "; } ?></p>
        </div>
  <?php if (count($down_hosts) > 0) { ?>
      <table id="broken_hosts" class="widetable">
      <tr><th>Hostname</th><th width="150px">State</th><th>Duration</th><th>Attempts</th><th>Detail</th></tr>
  <?php
      foreach($down_hosts as $host) {
          $controls = build_controls($host['tag'], $host['hostname'], '');
          echo "<tr id='host_row' class='{$nagios_host_status_colour[$host['host_state']]}'>";
          echo "<td>{$host['hostname']} " . print_tag($host['tag']) . " <span class='controls'>{$controls}</span></td>";
          echo "<td><blink>{$nagios_host_status[$host['host_state']]}</blink></td>";
          echo "<td>{$host['duration']}</td>";
          echo "<td>{$host['current_attempt']}/{$host['max_attempts']}</td>";
          echo "<td class=\"desc\">{$host['detail']}</td>";
          echo "</tr>";
      }
  ?>
      </table>
  <?php } else { ?>
      <table class="widetable status_green"><tr><td><b>All hosts OK</b></td></tr></table>
  <?php
  }
}
if (count($known_hosts) > 0) {
    foreach ($known_hosts as $this_host) {
        if ($this_host['is_ack']) $status_text = "ack";
        if ($this_host['is_downtime']) $status_text = "downtime";
        if (!$this_host['is_enabled']) $status_text = "disabled";
        $known_host_list[] = "{$this_host['hostname']} " . print_tag($this_host['tag']) . " <span class='known_hosts_desc'>({$status_text} - {$this_host['duration']})</span>";
    }
    $known_host_list_complete = implode(" &bull; ", $known_host_list);
    echo "<table class='widetable known_hosts'><tr><td><b>Known Problem Hosts: </b> {$known_host_list_complete}</td></tr></table>";
}
?>

    </div>
</div>

<?php if (empty($service_summary)) { ?>
  <div class="frame">
      <div class="section">
        <div class="header">
          <h3>Service status</h3>
      </div>
      <table class="widetable status_red"><tr><td><b>No Services given</b></td></tr></table>
<?php
} else { ?>
  <div class="frame">
      <div class="section">
        <div class="header">
          <h3>Service status</h3>
          <p class="totals"><b>Total:</b> <?php foreach($service_summary as $state => $count) { echo "<span class='{$nagios_service_status_colour[$state]}'>{$count}</span> "; } ?></p>
      </div>
  <?php if (count($broken_services) > 0) { ?>
      <table class="widetable" id="broken_services">
      <tr><th width="30%">Hostname</th><th width="50%">Service</th><th width="10%">Duration</th><th width="5%">Attempt</th></tr>
  <?php
      if ($sort_by_time) {
          usort($broken_services,'cmp_last_state_change');
      }
  foreach($broken_services as $service) {
      if (ignore($service['service_name'])){
          $soft_style = ($service['is_hard']) ? "" : "status_soft";
          $blink_tag = ($service['is_hard'] && $enable_blinking) ? "<blink>" : "";
          $controls = build_controls($service['tag'], $service['hostname'], $service['service_name']);
          $attempts = $service['is_hard'] ? "HARD" : "{$service['current_attempt']}/{$service['max_attempts']}";
          echo "<tr>";
          echo "<td>{$service['hostname']} " . print_tag($service['tag']) . " <span class='controls'>{$controls}</span></td>";
          echo "<td class='bold {$nagios_service_status_colour[$service['service_state']]} {$soft_style}'>{$blink_tag}{$service['service_name']}<span class='detail'>{$service['detail']}</span></td>";
          echo "<td>{$service['duration']}</td>";
          echo "<td>{$attempts}</td>";
          echo "</tr>";
      }
  }
  ?>
      </table>
  <?php } else { ?>
      <table class="widetable status_green"><tr><td><b>All services OK</b></td></tr></table>
  <?php }
}

if ($sort_by_time) {
    usort($known_services,'cmp_last_state_change');
}
######### DISABLE KNOWN SERVICE DISPLAY ON REQUEST
if ($_SESSION['known'] != "on"){
  unset($known_services);
  $known_services = array();
}


if (count($known_services) > 0) { ?>
    <h4>Known Service Problems</h4>
    <table class="widetable known_service" id="known_services">
    <tr><th width="30%">Hostname</th><th width="37%">Service</th><th width="18%">State</th><th width="10%">Duration</th><th width="5%">Attempt</th></tr>
<?php

    foreach($known_services as $service) {
        if ($service['is_ack']) $status_text = "ack";
        if ($service['is_downtime']) $status_text = "downtime {$service['downtime_remaining']}";
        if (!$service['is_enabled']) $status_text = "disabled";
        echo "<tr class='known_service'>";
        echo "<td>{$service['hostname']} " . print_tag($service['tag']) . "</td>";
        echo "<td>{$service['service_name']}</td>";
        echo "<td class='{$nagios_service_status_colour[$service['service_state']]}'>{$nagios_service_status[$service['service_state']]} ({$status_text})</td>";
        echo "<td>{$service['duration']}</td>";
        echo "<td>{$service['current_attempt']}/{$service['max_attempts']}</td>";
        echo "</tr>";
    }
?>

    </table>
<?php } ?>

    </div>
</div>

<?php

echo "<!-- nagios-api server status: -->";
foreach ($curl_stats as $server => $server_stats) {
    echo "<!-- {$server_stats['url']} returned code {$server_stats['http_code']}, {$server_stats['size_download']} bytes ";
    echo "in {$server_stats['total_time']} seconds (first byte: {$server_stats['starttransfer_time']}). JSON parsed {$server_stats['objects']} hosts -->\n";
}

?>
<?php


// Utility function to sort the aggregated array by keys.
function deep_ksort(&$arr) {
    ksort($arr);
    foreach ($arr as &$a) {
        if (is_array($a) && !empty($a)) {
            deep_ksort($a);
        }
    }
}

function cmp_last_state_change($a,$b) {
    if ($a['last_state_change'] == $b['last_state_change']) return 0;
    return ($a['last_state_change'] > $b['last_state_change']) ? -1 : 1;
}

function build_controls($tag, $host, $service) {
    $tag = implode(',', $tag);
    $controls = '<div class="btn-group">';
    $controls .= "<a href='#' onClick=\"$.post('do_action.php', {
        nag_host: '{$tag}', hostname: '{$host}', service: '{$service}', action: 'ack' }, function(data) { showInfo(data) } ); return false;\" class='btn btn-mini'>
            <i class='icon-check'></i> Ack </a>";
    if (!isset($service['is_enabled'])) {
        $controls .="<a href='#' onClick=\"$.post('do_action.php', {
                nag_host: '{$tag}', hostname: '{$host}', service: '{$service}', action: 'disable' }, function(data) { showInfo(data) } ); return false;\" class='btn btn-mini'>
                    <i class='icon-volume-off'></i> Silence</a>";
    } else {
        $controls .="<a href='#' onClick=\"$.post('do_action.php', {
                nag_host: '{$tag}', hostname: '{$host}', service: '{$service}', action: 'enable' }, function(data) { showInfo(data) } ); return false;\" class='btn btn-mini'>
                    <i class='icon-volume-up'></i> Unsilence</a>";
    }
    $controls .= '
        <a class="btn btn-mini dropdown-toggle" data-toggle="dropdown" href="#">
        <i class="icon-time"></i> Downtime <span class="caret"></span></a>
        <ul class="dropdown-menu pull-right">';
        $timespans = array("10 minutes" => 10, "30 minutes" => 30, "60 minutes" => 60, "2 hours" => 120, "12 hours" => 720, "1 day" => 1440, "7 days" => 10080);
        foreach ($timespans as $name => $minutes) {
            $controls .= "<li><a onClick=\"$.post('do_action.php',
                { nag_host: '{$tag}', hostname: '{$host}', service: '{$service}', duration: {$minutes}, action: 'downtime' }, function(data) { showInfo(data) } ); return false;\"
                href='#'>{$name}</a></li>";
        }
        $controls .= "</ul>";
    $controls .= "</div>";
    return $controls;
}
function curl_get_file_contents($URL)
    {
        $c = curl_init();
        curl_setopt($c, CURLOPT_RETURNTRANSFER, 1);
        curl_setopt($c, CURLOPT_URL, $URL);
        $contents = curl_exec($c);
        curl_close($c);

        if ($contents) return $contents;
            else return FALSE;
    }
#$ddate = file_get_contents('http://api.ddate.cc/v1/today.txt');
echo "<center><h3>";
if ( file_exists('/var/www/dash/Nagdash/ddate.txt')) {
  $ddate = file_get_contents('/var/www/dash/Nagdash/ddate.txt');
  echo "$ddate";
} else {
  if ( file_exists('/usr/bin/ddate')) {
    $ddate = exec("/usr/bin/ddate");
    echo "$ddate";
  } else {
    echo "ddate not reachable!";
  }
}
echo "</h3><br><h3>";#
if (file_exists('temp.txt')) {
  include("temp.txt");
  echo "°C";
} else {
  echo "Temperature not reachable!";
}
echo "</h3></center>";

//$biertime = "not sure if its biertime…";
$beer_url = "http://bier.synyx.coffee";
$beertime = curl_get_file_contents($beer_url);
if ( $beertime != FALSE ) {
  if (preg_match('/<title>(.+)<\/title>/',file_get_contents($beer_url),$matches) && isset($matches[1])){
    $biertime = $matches[1];
  } else{
    $biertime = "not sure if its biertime…";
  }
} else {
  $biertime = "biertime not reachable…";
}
#echo "<center><h3>$biertime</h3></center><br>".date("R");
echo "<center><h3>$biertime</h3></center><br>";

include("synyx.html");

?>
