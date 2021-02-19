<?php

/**
 * @param $url
 * @param array $options
 * @return string | array Error as string
 */
function connectGitlabMRs($url, $options = array()) {
  $default_options = array(
    'wip' => 'no',
    'state' => 'opened',
    'order_by' => 'updated_at',
    'sort' => 'desc',
    'scope' => 'all',
    'target_branch' => 'master',
  );
  $options = array_merge($default_options, $options);
  $state = gitlabV4Get($url.'/merge_requests', $options);

  if (is_string($state)) {
    return $state;
  }

  if (count($state) == 0) {
    return array("url" => array(
      'services' => [],
      'downtimes' => [],
      'current_state' => 1,
    ));
  }

  $host_state = array_reduce($state, function ($hosts, $mr) {
    $hn = explode('!', $mr['references']['full'])[0];
    if (!isset($hosts[$hn])) {
      $hosts[$hn] = array(
        'services' => [],
        'downtimes' => [],
        'current_state' => 0, # "host" is always up
      );
    }

    $startsAt = DateTime::createFromFormat('Y-m-d\TH:i:s+', $mr['updated_at'],  new DateTimeZone('Etc/Zulu'));
    $description = 'Author: '.$mr['author']['name'].($mr['assignee'] ? 'Assigned To: '.$mr['assignee']['name'] : '');
    $link = '<a href="'.$mr['web_url'].'" target="_blank">MR '.$mr['references']['short'].'</a>';

    $sn = 'MR '.$mr['references']['short'].': '.$mr['title'];
    $hosts[$hn]['services'][$sn] = array(
      'current_state'                 => 1,
      'problem_has_been_acknowledged' => 0,
      'scheduled_downtime_depth'      => 0,
      'notifications_enabled'         => 1,
      'plugin_output'                 => $description . ' ' . $link,
      'max_attempts'                  => 1,
      'current_attempt'               => 1,
      'state_type'                    => 1,
      'downtimes'                     => [],
      'last_state_change'             => $startsAt->getTimestamp(),
      'labels'                        => $mr['labels'],
    );

    return $hosts;
  }, array());

  return $host_state;
}

function gitlabV4Get($url, $params = array()) {
  global $curl_stats;

  $hostname    = parse_url($url, PHP_URL_HOST);
  $port        = parse_url($url, PHP_URL_PORT);
  $token       = parse_url($url, PHP_URL_PASS);
  $request_url = "$url";
  $headers     = [
    'Accept: application/json',
    'X-HTTP-Method-Override: GET',
    'Authorization: Bearer ' . $token,
  ];
  $ch          = curl_init();
  curl_setopt_array(
    $ch, [
      CURLOPT_URL            => $request_url,
      CURLOPT_HTTPHEADER     => $headers,
      CURLOPT_RETURNTRANSFER => true,
      CURLOPT_SSL_VERIFYHOST => 0,
      CURLOPT_POSTFIELDS     => http_build_query($params),
      CURLOPT_CUSTOMREQUEST  => 'GET',
    ]
  );
  if (!$json = curl_exec($ch)) {
    return "<pre>Attempt to hit GitLab API failed, sorry. Curl said: " . curl_error($ch) . "</pre>";
  } else {
    $curl_stats["$hostname:$port"] = curl_getinfo($ch);
  }
  curl_close($ch);

  if (!$state = json_decode($json, true)) {
    return "Attempt to parse gitlab json failed, sorry (JSON decode failed)";
  }
  $curl_stats["$hostname:$port"]['objects'] += count($state);

  return $state;
}
