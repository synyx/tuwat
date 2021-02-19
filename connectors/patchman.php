<?php

/**
 * @param $url
 * @return string | array Error as string
 */
function connectPatchman($url) {
  $username    = parse_url($url, PHP_URL_USER);
  $password    = parse_url($url, PHP_URL_PASS);

  $hosts = patchmanGet($url . "/api/host/", $username, $password);
  if (is_string($hosts)) {
    return $hosts;
  }

  if (array_key_exists('next', $hosts)) {
    $all_hosts = $hosts['results'];
    while ($hosts['next']) {
      $hosts = patchmanGet($hosts['next'], $username, $password);
      if (is_string($hosts)) {
        return $hosts;
      }
      array_push($all_hosts, $hosts['results']);
    }
  } else {
    $all_hosts = $hosts;
  }

  return patchmanHosts2Nagios($all_hosts);
}

/**
 * @param $hosts array
 * @return array|string Error as string
 */
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

    if ($host['security_update_count'] > 0 || $host['bugfix_update_count'] > 25) {
      $sn                          = 'Patch level insufficient';
      $description = "Security updates: {$host['security_update_count']}, Updates: {$host['bugfix_update_count']}";
      $hosts[$hn]['downtimes'] = [];
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
  #return array(array_key_first($host_state) => $host_state[array_key_first($host_state)]);
  return $host_state;
}

function patchmanGet($url, $username, $password) {
  $hostname    = parse_url($url, PHP_URL_HOST);
  $path    = parse_url($url, PHP_URL_PATH);
  $port        = parse_url($url, PHP_URL_PORT);
  $params        = parse_url($url, PHP_URL_QUERY);
  $request_url = "https://{$hostname}{$path}?{$params}";
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
      CURLOPT_USERPWD        => $username . ":" . $password,
      CURLOPT_POSTFIELDS     => $params,
      CURLOPT_CUSTOMREQUEST  => 'GET',
      CURLOPT_VERBOSE        => true,
    ]
  );
  if (!$response = curl_exec($ch)) {
    return "<pre>Attempt to hit patchman failed, sorry. Curl said: " . curl_error($ch) . $request_url . "</pre>";
  } else {
    $curl_stats["$hostname:$port"] = curl_getinfo($ch);
  }
  $header_size = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
  curl_close($ch);
  #$header = substr($response, 0, $header_size);
  $body = substr($response, $header_size);

  if (!$state = json_decode($body, true)) {
    return "Attempt to parse gitlab json failed, sorry (JSON decode failed): $body";
  }
  $curl_stats["$hostname:$port"]['objects'] += count($state);

  #print_r($header);
  return $state;
}
