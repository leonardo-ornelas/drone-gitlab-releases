# Drone gitlab releases

## The Gitlab release plugin create releases. The below pipeline configuration demonstrates simple usage:

```yaml
kind: pipeline
name: default

steps:
- name: release
  image: solutisdigital/drone-gitlab-releases
  settings:
    token: authtoken
    assets: asset.zip
```

## Example configuration using credentials from secrets:

```yaml
steps:
- name: release
  image: solutisdigital/drone-gitlab-releases
  settings:
    assets: asset.zip
    token:
      from_secret: gitlab_token
```