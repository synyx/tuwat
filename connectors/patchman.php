<?php

/**
 * @param $url
 * @return string | array Error as string
 */
function connectPatchman($url) {
  #return apcu_entry("patchman-{$url}", 'patchmanHosts', 3600);
  return patchmanHosts($url);
}

/**
 * @param $url
 *
 * @return array|string Error as string
 */
function patchmanHosts($url) {
  $result = patchmanApiGet($url, "host");

  $page = 1;
  do {
    $result = patchmanApiGet($url, "host", array('page'=>$page));
    if (is_string($result)) {
      return $result;
    }
    $hosts = $result['results'];
    $page++;
  } while ($result['next']);

  if (count($hosts) == 0) {
    return array("url" => array(
      'services' => [],
      'downtimes' => [],
      'current_state' => 1,
    ));
  }

  $host_state = array_reduce($hosts, function ($hosts, $host) {
    $hn = $host['hostname'];
    $labels = $host['tags'];
    if (!isset($hosts[$hn])) {
      $hosts[$hn] = array(
        'services' => [],
        'downtimes' => [],
        'current_state' => 0, # "host" is always up
      );
    }

    $startsAt = DateTime::createFromFormat('Y-m-d\TH:i:s+', $host['lastreport'],  new DateTimeZone('Etc/Zulu'));
    $description = '';

    $sn = 'Patch level insufficient';
    $hosts[$hn]['services'][$sn] = array(
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
      'labels'                        => $labels,
    );

    return $hosts;
  }, array());

  #print_r($host_state);
  return array(array_key_first($host_state) => $host_state[array_key_first($host_state)]);
}


/**
 * Query Patchman API.
 *
 * @param   string  $url       to query.
 * @param   string  $endpoint  to query data from.
 * @param   array   $params    for request
 *
 * @return array|mixed
 */
function patchmanApiGet(string $url, string $endpoint, $params = array()) {
  $hostname    = parse_url($url, PHP_URL_HOST);
  $port        = parse_url($url, PHP_URL_PORT);
  $username    = parse_url($url, PHP_URL_USER);
  $password    = parse_url($url, PHP_URL_PASS);
  $request_url = "$url/api/{$endpoint}/";
  $headers     = [
    'Accept: application/json',
    'X-HTTP-Method-Override: GET',
  ];
  $ch          = curl_init();
  curl_setopt_array(
    $ch, [
      CURLOPT_URL            => $request_url,
      CURLOPT_HTTPHEADER     => $headers,
      CURLOPT_USERPWD        => $username . ":" . $password,
      CURLOPT_RETURNTRANSFER => true,
      #CURLOPT_SSL_VERIFYHOST => 2,
      #CURLOPT_SSL_VERIFYPEER => 1
      CURLOPT_SSL_VERIFYHOST => 0,
      CURLOPT_SSL_VERIFYPEER => 0,
      CURLOPT_POSTFIELDS     => http_build_query($params),
      CURLOPT_CUSTOMREQUEST  => 'GET',
    ]
  );
  if (!$json = curl_exec($ch)) {
    return ["<pre>Attempt to hit patchman API failed, sorry. Curl said: " . curl_error($ch) . "</pre>"];
  } else {
    $curl_stats["$hostname:$port"] = curl_getinfo($ch);
  }
  curl_close($ch);

  if (!$state = json_decode($json, true)) {
    return ["Attempt to hit patchman API failed, sorry (JSON decode failed)"];
  }
  $curl_stats["$hostname:$port"]['objects'] += count($state['results']);

  return $state;
}
