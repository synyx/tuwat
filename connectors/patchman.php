<?php

/**
 * @param $url
 * @return string | array Error as string
 */
function connectPatchman($url) {
  #return apcu_entry("patchman-{$url}", 'patchmanHosts', 3600);
  return patchmanLogin($url);
}

function patchmanHosts2Nagios($hosts) {

  if (count($hosts) == 0) {
    return array("url" => array(
      'services' => [],
      'downtimes' => [],
      'current_state' => 1,
    ));
  }

  $host_state = array_reduce($hosts, function ($hosts, $host) {
    $hn = $host['hostname'];

    $startsAt = DateTime::createFromFormat('Y-m-d\TH:i:s+', $host['lastreport'],  new DateTimeZone('Etc/Zulu'));

    if ($host['security'] > 0) {
      $sn                          = 'Patch level insufficient';
      $description = "Security updates: {$host['security']}, Updates: {$host['updates']}";
      $hosts[$hn]['services'][$sn] = [
        'current_state'                 => 1,
        'problem_has_been_acknowledged' => 0,
        'scheduled_downtime_depth'      => 0,
        'notifications_enabled'         => 1,
        'plugin_output'                 => $description,
        'max_attempts'                  => 1,
        'current_attempt'               => 1,
        'state_type'                    => 1,
        'downtimes'                     => [],
        'last_state_change'             => $startsAt->getTimestamp(),
        'labels'                        => '',
      ];
    }

    return $hosts;
  }, array());

  #print_r($host_state);
  return array(array_key_first($host_state) => $host_state[array_key_first($host_state)]);
}

function patchmanLogin(string $url) {
  $html = patchmanGet($url . '/login/');
  $doc = new DOMDocument();
  $doc->loadHTML($html);
  $inputs = $doc->getElementsByTagName('input');

  $csrftoken = null;
  foreach ($inputs as $input) {
    foreach ($input->attributes as $attr) {
      if ($attr->nodeValue == 'csrfmiddlewaretoken') {
        foreach ($input->attributes as $vattr) {
          if ($vattr->nodeName == 'value') {
            $csrftoken = $vattr->nodeValue;
          }
        }
      }
    }
  }

  if ($csrftoken) {
    $username    = parse_url($url, PHP_URL_USER);
    $password    = parse_url($url, PHP_URL_PASS);

    $data = [
      'csrfmiddlewaretoken' => $csrftoken,
      'username' => $username,
      'password' => $password,
      'next' => '/dashboard/',
    ];
    patchmanPost($url . '/login/', $data);

    $hosts = array();

    $html = patchmanGet($url . '/dashboard/');
    $doc = new DOMDocument();
    $doc->loadHTML($html);

    $updates_node = $doc->getElementById('bugupdate_hosts');
    $updates = $updates_node ? $updates_node->getElementsByTagName('tr') : [];
    foreach ($updates as $update) {
      [$hostname, $nr] = patchmanExtractUpdatesFromNode($update, 2);
      if (!isset($hosts[$hostname])) $hosts[$hostname] = array('hostname' => $hostname);
      $hosts[$hostname]['updates'] = intval($nr);
    }

    $secupdates_node = $doc->getElementById('secupdate_hosts');
    $secupdates = $secupdates_node ? $secupdates_node->getElementsByTagName('tr') : [];
    foreach ($secupdates as $update) {
      [$hostname, $nr] = patchmanExtractUpdatesFromNode($update, 1);
      if (!isset($hosts[$hostname])) $hosts[$hostname] = array('hostname' => $hostname);
      $hosts[$hostname]['security'] = intval($nr);
    }

    return $hosts;
  }
  return [];
}

function patchmanExtractUpdatesFromNode(DOMElement $node, int $column) {
  $tds = $node->getElementsByTagName('td');
  $host = $tds->item(0)->textContent;
  $nr = $tds->item($column)->textContent;
  $lastreport = $tds->item(5)->textContent;

  return [$host, $nr, $lastreport];
}

function patchmanGet(string $url, $params = array()) {
  $hostname    = parse_url($url, PHP_URL_HOST);
  $port        = parse_url($url, PHP_URL_PORT);
  $request_url = $url;
  $headers     = [
    'Accept: */*',
    'X-HTTP-Method-Override: GET',
  ];
  $ch          = curl_init();
  curl_setopt_array(
    $ch, [
      CURLOPT_URL            => $request_url,
      CURLOPT_HTTPHEADER     => $headers,
      CURLOPT_RETURNTRANSFER => true,
      CURLOPT_HEADER         => true,
      #CURLOPT_SSL_VERIFYHOST => 2,
      #CURLOPT_SSL_VERIFYPEER => 1
      CURLOPT_SSL_VERIFYHOST => 0,
      CURLOPT_SSL_VERIFYPEER => 0,
      CURLOPT_POSTFIELDS     => http_build_query($params),
      CURLOPT_CUSTOMREQUEST  => 'GET',
      CURLOPT_COOKIEJAR      => '/tmp/nagdash_cookiejar.txt',
      CURLOPT_COOKIEFILE     => '/tmp/nagdash_cookiejar.txt',
      CURLOPT_VERBOSE        => true,
    ]
  );
  if (!$response = curl_exec($ch)) {
    return ["<pre>Attempt to hit patchman failed, sorry. Curl said: " . curl_error($ch) . "</pre>"];
  } else {
    $curl_stats["$hostname:$port"] = curl_getinfo($ch);
  }
  $header_size = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
  curl_close($ch);
  $header = substr($response, 0, $header_size);
  $body = substr($response, $header_size);
  #print_r($header);
  return $body;
}


function patchmanPost(string $url, array $payload) {
  $hostname    = parse_url($url, PHP_URL_HOST);
  $port        = parse_url($url, PHP_URL_PORT);
  $request_url = $url;
  $headers     = [
    'Accept: */*',
  ];
  $ch          = curl_init();
  curl_setopt_array(
    $ch, [
      CURLOPT_URL            => $request_url,
      CURLOPT_HTTPHEADER     => $headers,
      CURLOPT_RETURNTRANSFER => true,
      CURLOPT_HEADER         => true,
      #CURLOPT_SSL_VERIFYHOST => 2,
      #CURLOPT_SSL_VERIFYPEER => 1
      CURLOPT_SSL_VERIFYHOST => 0,
      CURLOPT_SSL_VERIFYPEER => 0,
      CURLOPT_POSTFIELDS     => http_build_query($payload),
      CURLOPT_CUSTOMREQUEST  => 'POST',
      CURLOPT_COOKIEFILE     => '/tmp/nagdash_cookiejar.txt',
      CURLOPT_COOKIEJAR      => '/tmp/nagdash_cookiejar.txt',
      CURLOPT_VERBOSE        => true,
    ]
  );
  if (!$response = curl_exec($ch)) {
    return ["<pre>Attempt to post to patchman failed, sorry. Curl said: " . curl_error($ch) . "</pre>"];
  } else {
    $curl_stats["$hostname:$port"] = curl_getinfo($ch);
  }
  $header_size = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
  curl_close($ch);
  $header = substr($response, 0, $header_size);
  $body = substr($response, $header_size);
  #print_r($header);
  return $body;
}
