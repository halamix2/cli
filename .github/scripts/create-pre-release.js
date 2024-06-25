module.exports = async function ({ github, context }) {
  const fs = require("fs");

  const org = context.payload.repository.owner.login;
  const repo = context.payload.repository.name;
  const tag = "dev";
  const fileName = "kyma-test";
  const filePath = "bin/kyma-test";
  const filedata = fs.readFileSync(filePath);

  var release = null;

  try {
    release = await github.rest.repos.getReleaseByTag({
      owner: org,
      repo: repo,
      tag: tag,
    });
    // update release body
    await github.rest.repos.updateRelease({
      owner: org,
      repo: repo,
      release_id: release.data.id,
      tag_name: "dev",
      name: "dev release",
      body: "This is latest release from " + context.sha + " commit",
      prerelease: true,
    });
  } catch (error) {
    // if a release doesn't exist, it throws a 404 error
    if (error.status && error.status === 404) {
      console.log("creating release object");
      release = await github.rest.repos.createRelease({
        owner: org,
        repo: repo,
        tag_name: "dev",
        name: "dev release",
        body: "This is latest release from " + context.sha + " commit",
        prerelease: true,
      });
    } else {
      console.error("failed to find release object", error);
      throw new Error("failed to find release object");
    }
  }

  // check if file exists
  var { data: existingReleaseAssets } =
    await github.rest.repos.listReleaseAssets({
      owner: org,
      repo: repo,
      release_id: release.data.id,
    });

  console.log(
    "existing assets for the release: ",
    existingReleaseAssets.flatMap((asset) => asset.name)
  );

  var foundFile = existingReleaseAssets.find(
    (asset) => asset.name === fileName
  );

  if (!foundFile) {
    console.log("file not found, creating");
    await github.rest.repos.uploadReleaseAsset({
      owner: org,
      repo: repo,
      release_id: release.data.id,
      name: fileName,
      data: filedata,
      headers: {
        "content-type": "text/plain",
        "content-length": filedata.length,
      },
    });
  } else {
    console.log("file found, updating");
    await github.rest.repos.deleteReleaseAsset({
      owner: org,
      repo: repo,
      asset_id: foundFile.id,
    });
    await github.rest.repos.uploadReleaseAsset({
      owner: org,
      repo: repo,
      release_id: release.data.id,
      name: fileName,
      data: filedata,
      headers: {
        "content-type": "text/plain",
        "content-length": filedata.length,
      },
    });
  }
};
